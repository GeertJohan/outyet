package main

import (
	"github.com/jessevdk/go-flags"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	expectingVersion = "1.1.1 DISABLED"                                  // number being expected. must be changed manually (for now).
	changeURLBase    = "https://code.google.com/p/go/source/detail?r=go" // base url to poll the tag
	updateInterval   = 6 * time.Second                                   // Update interval for the expected number
)

var defaultPage = "http://isgo1point3.outyet.org"

var (
	versions     = make(map[string]*version) // map with all versions by number(string)
	versionsLock sync.RWMutex                // map lock
)

var regexpNumber = regexp.MustCompile(`^[1-9](?:\.[0-9]){0,2}$`)

var colVersions *mgo.Collection
var colNV *mgo.Collection

var options struct {
	Listen string `short:"l" long:"listen" default:"141.138.139.6:80" description:"IP:post to listen on"`
}

func main() {
	args, err := flags.Parse(&options)
	if err != nil {
		log.Fatalln(err)
	}
	if len(args) > 0 {
		log.Fatalln("Unexpected arguments.")
	}

	mgoSess, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalln(err)
	}
	colVersions = mgoSess.DB("outyet").C("versions")
	colVersions.EnsureIndex(mgo.Index{
		Key:      []string{"number"},
		Unique:   true,
		DropDups: true,
	})
	colNV = mgoSess.DB("outyet").C("namevalue")
	colNV.EnsureIndex(mgo.Index{
		Key:      []string{"name"},
		Unique:   true,
		DropDups: true,
	})

	if err := http.ListenAndServe(options.Listen, http.HandlerFunc(rootHandler)); err != nil {
		log.Fatalln(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// handler for stats page
	if r.Host == "stats.outyet.org" {
		statsHandler(w, r)
		return
	}

	// redirect for 'old' domain
	if r.Host == "isgo1point2outyet.com" {
		http.Redirect(w, r, "http://isgo1point2.outyet.org", http.StatusTemporaryRedirect)
		return
	}

	// only handle requests on /
	if r.RequestURI != "/" {
		http.NotFound(w, r)
		return
	}

	// check if Host header matches isgo*.outyet.org
	if !strings.HasSuffix(r.Host, ".outyet.org") || !strings.HasPrefix(r.Host, "isgo") {
		log.Printf("Invalid host format detected. %s\n", r.Host)
		http.Redirect(w, r, defaultPage, http.StatusTemporaryRedirect)
		return
	}

	number := strings.Replace(r.Host[4:len(r.Host)-11], "point", ".", -1)
	log.Println(number)

	if !regexpNumber.MatchString(number) {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if strings.HasSuffix(number, ".0") {
		number = number[:len(number)-2]
		if len(number) > 0 {
			http.Redirect(w, r, "http://isgo"+strings.Replace(number, ".", "point", -1)+".outyet.org", code)
			return
		}
		http.Redirect(w, r, defaultPage, http.StatusTemporaryRedirect)
		return
	}

	// get right version in a safe way
	o := getVersion(number)

	// add hitCount's
	colVersions.Upsert(bson.M{"number": o.number}, bson.M{"$inc": bson.M{"hits": 1}})
	colNV.Upsert(bson.M{"name": "counts"}, bson.M{"$inc": bson.M{"hits": 1}})

	// execute template
	data := dataOutyet{
		Outyet: <-o.isOutyetChan, //retrieve outyet directly from channel
		Number: number,
	}
	err := tmplOutyet.Execute(w, data)
	if err != nil {
		log.Print(err)
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	data := &dataStats{}

	colNV.Find(bson.M{"name": "counts"}).One(data)
	colVersions.Find(nil).Sort("number").All(&data.Versions)

	for _, v := range data.Versions {
		// get outyet for given version number
		v.Outyet = <-getVersion(v.Number).isOutyetChan

		// add hitCount's
		colVersions.Upsert(bson.M{"number": v.Number}, bson.M{"$inc": bson.M{"hits": 1}})
	}

	err := tmplStats.Execute(w, data)
	if err != nil {
		log.Print(err)
	}
}
