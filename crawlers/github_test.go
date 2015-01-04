// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crawlers

import "testing"

func TestIsLanguageWanted(t *testing.T) {
	wantedLangs := []string{"go", "ruby", "java"}
	prjLangs := map[string]int{
		"JavaScript": 434332,
		"Go":         343,
		"HTML":       3432,
	}

	// expect true
	if ok, err := isLanguageWanted(wantedLangs, prjLangs); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Errorf("isLanguageWanted(%v, %v): expected 'true' found 'false'",
			wantedLangs, prjLangs)
	}

	prjLangs = map[string]int{
		"JavaScript": 434332,
		"Python":     343,
		"HTML":       3432,
	}

	// expect false
	if ok, err := isLanguageWanted(wantedLangs, prjLangs); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Errorf("isLanguageWanted(%v, %v): expected 'false' found 'true'",
			wantedLangs, prjLangs)
	}
}
