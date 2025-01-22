package tagger

import (
	"log"
	"path/filepath"

	"9fans.net/go/acme"
)

func ExecuteLogLoop(config Configuration) {
	acmeLog, err := acme.Log()
	if err != nil {
		log.Printf("could not create log watcher: %v\n", err)
		return
	}
	defer acmeLog.Close()

	for {
		event, err := acmeLog.Read()
		if err != nil {
			log.Printf("could not read acme log: %v\n", err)
			return
		}

		switch event.Op {
		case "new", "put":
			ext := filepath.Ext(event.Name)
			tc := config.Find(ext)

			if tc != nil {
				maybeTagWindow(tc, event.ID, event.Name)
			}

		default:
			// log.Printf("not interested in %#v\n", event)
		}
	}
}
