package main

import (
	"expvar"
	"log"
	"net/http"
	"time"
)

// defines a number to check
type version struct {
	number          string         // version number to check
	isOutyetChan    chan bool      // chan that tells wether the version has been tagged
	lastCheck       time.Time      // last time this version was checked
	hitCount        *expvar.Int    // total times this version was viewed
	checkCount      *expvar.Int    // total times this version was checked
	checkError      *expvar.String // last check error for this version
	checkErrorCount *expvar.Int    // check error count for this version
}

func (vers *version) run() {
	var status bool

ReCheck:
	for {
		vers.checkCount.Add(1) // HL
		totalCheckCount.Add(1) //HL
		r, err := http.Head(changeURLBase + vers.number)
		if err != nil {
			log.Print(err)
			vers.checkError.Set(err.Error()) // HL
			vers.checkErrorCount.Add(1)      // HL
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
		number:          number,
		isOutyetChan:    make(chan bool),
		hitCount:        expvar.NewInt("hitCount" + number),
		checkCount:      expvar.NewInt("checkCount" + number),
		checkError:      expvar.NewString("pollError" + number),
		checkErrorCount: expvar.NewInt("pollErrorCount" + number),
	}
	go o.run()
	versions[o.number] = o
	// all done
	versionsLock.Unlock()
	return o
}
