package tagger

import (
	"fmt"
	"strings"

	"9fans.net/go/acme"

	"github.com/jcowgar/acme-utils/internal/config"
)

func hasTag(win *acme.Win, tag string) (bool, error) {
	tagBytes := make([]byte, 256)
	_, err := win.Read("tag", tagBytes)
	if err != nil {
		return false, fmt.Errorf("could not read tag of winID: %d: %v\n",
			win.ID(), err,
		)
	}

	return strings.Contains(string(tagBytes), tag), nil
}

func maybeTagWindow(tc *config.TagConfig, winID int, filename string) error {
	win, err := acme.Open(winID, nil)
	if err != nil {
		return fmt.Errorf("could not open winID %d: %v\n", winID, err)
	}

	alreadyHasTag, err := hasTag(win, tc.Tag)
	if err != nil {
		return fmt.Errorf("could not check for winID: %d: %v\n", winID, err)
	}
	if alreadyHasTag {
		return nil
	}

	_, err = win.Write("tag", []byte(" "+tc.Tag))
	if err != nil {
		return fmt.Errorf("could not write tag of winID: %d: %v\n",
			winID, err)
	}

	return nil
}
