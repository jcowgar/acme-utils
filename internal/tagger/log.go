package tagger

import (
	"log"
	"path/filepath"
	"strings"

	"9fans.net/go/acme"

	"github.com/jcowgar/acme-utils/internal/config"
)

func ExecuteLogLoop(config config.Config) {
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

		if event.Name == "" {
			continue
		}

		switch event.Op {
		case "new":
			maybe_tag(config, event.ID, event.Name)

		case "put":
			maybe_tag(config, event.ID, event.Name)
			maybe_reformat(config, event.ID, event.Name)

		default:
			// log.Printf("not interested in %#v\n", event)
		}
	}
}

func maybe_tag(config config.Config, winID int, filename string) {
	ext := filepath.Ext(filename)
	tc := config.TaggerFor(ext)

	if tc != nil {
		maybeTagWindow(tc, winID, filename)
	}
}

func maybe_reformat(config config.Config, winID int, filename string) {
	ext := filepath.Ext(filename)
	formatter := config.FormatterFor(ext)
	if formatter == nil {
		return
	}

	// First split the command string into parts
	parts := strings.Fields(formatter.Command)

	// Create a result slice
	cmd := make([]string, len(parts))

	// Replace %f with filename in each part
	for i, part := range parts {
		cmd[i] = strings.ReplaceAll(part, "%f", filename)
	}

	reformat_win(winID, cmd, filename, formatter.InPlace)
}
