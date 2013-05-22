package main

import (
	"expvar"
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	expectingVersion = "1.1.1"                                           // number being expected. must be changed manually (for now).
	changeURLBase    = "https://code.google.com/p/go/source/detail?r=go" // base url to poll the tag
	updateInterval   = 6 * time.Second                                   // Update interval for the expected number
)

var defaultPage = "isgo" + strings.Replace(expectingVersion, ".", "point", -1) + ".outyet.org" //++ TODO(GeertJohan): strings replace "." to "point" ?

var (
	totalHitCount   = expvar.NewInt("hitCountTotal")   // total amount of hits
	totalCheckCount = expvar.NewInt("checkCountTotal") // total amount of checks
)

var (
	versions     = make(map[string]*version) // map with all versions by number(string)
	versionsLock sync.RWMutex                // map lock
)

func main() {
	flag.Parse()

	http.HandleFunc("/", rootHandler)
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatalln(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// redirect for 'old' domain
	if r.Host == "isgo1point2outyet.com" {
		http.Redirect(w, r, "isgo1point2.outyet.org", http.StatusTemporaryRedirect)
		return
	}

	if !strings.HasSuffix(r.Host, ".outyet.org") || !strings.HasPrefix("isgo", r.Host) {
		http.Redirect(w, r, defaultPage, http.StatusTemporaryRedirect)
		return
	}

	numberFromHost := r.Host[4 : len(r.Header)-11]
	log.Println(numberFromHost)

	number := expectingVersion

	// get right version in a safe way
	o := getVersion(number)

	// add hitCount's
	totalHitCount.Add(1) // HL
	o.hitCount.Add(1)    //HL

	// execute template
	data := dataOutyet{
		Outyet:  <-o.isOutyetChan, //retrieve outyet directly from channel
		Version: number,
	}
	err := tmplOutyet.Execute(w, data)
	if err != nil {
		log.Print(err)
	}
}
