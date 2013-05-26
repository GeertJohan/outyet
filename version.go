package main

import (
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"time"
)

// defines a number to check
type version struct {
	number       string    // version number to check
	isOutyetChan chan bool // chan that tells wether the version has been tagged
	lastCheck    time.Time // last time this version was checked
}

func (vers *version) run() {
	var status bool

ReCheck:
	for {
		colVersions.Upsert(bson.M{"number": vers.number}, bson.M{"$inc": bson.M{"checks": 1}})
		colNV.Upsert(bson.M{"name": "counts"}, bson.M{"$inc": bson.M{"checks": 1}})
		r, err := http.Head(changeURLBase + vers.number)
		if err != nil {
			log.Print(err)
			colVersions.Upsert(bson.M{"number": vers.number}, bson.M{"$push": bson.M{"errors": err.Error()}})
			status = false
		} else {
			status = r.StatusCode == http.StatusOK
		}
		vers.lastCheck = time.Now()

		// if version is out yet, fill isOutyetChan with that status forever!
		if status {
			for {
				vers.isOutyetChan <- status
			}
		}

		// depending on wether the version is expected, have automatically updating or user triggered updating.
		if vers.number == expectingVersion {
			// update automatically every `updateInterval`
			for {
				select {
				case vers.isOutyetChan <- status:
				case <-time.After(updateInterval):
					continue ReCheck
				}
			}
		} else {
			// update on user trigger and `updateInterval` timeout
			for {
				vers.isOutyetChan <- status
				if time.Now().Sub(vers.lastCheck) > updateInterval {
					continue ReCheck
				}
			}
		}
	}
}

func getVersion(number string) *version {
	// find version with readlock
	versionsLock.RLock()
	if o, exists := versions[number]; exists {
		versionsLock.RUnlock()
		return o
	}
	versionsLock.RUnlock()

	// version not found, let's create it
	versionsLock.Lock()
	// check again this time with write lock
	if o, exists := versions[number]; exists {
		versionsLock.Unlock()
		return o
	}
	// create and insert version
	o := &version{
		number:       number,
		isOutyetChan: make(chan bool),
	}
	colVersions.Insert(bson.M{"number": o.number, "createTime": time.Now()})
	go o.run()
	versions[o.number] = o
	// all done
	versionsLock.Unlock()
	return o
}
