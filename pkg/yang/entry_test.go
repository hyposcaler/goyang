// Copyright 2015 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yang

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNilEntry(t *testing.T) {
	e := ToEntry(nil)
	_, ok := e.Node.(*ErrorNode)
	if !ok {
		t.Fatalf("ToEntry(nil) did not return an error node")
	}
	errs := e.GetErrors()
	switch len(errs) {
	case 0:
		t.Fatalf("GetErrors returned no error")
	default:
		t.Errorf("got %d errors, wanted 1", len(errs))
		fallthrough
	case 1:
		got := errs[0].Error()
		want := "ToEntry called with nil"
		if got != want {
			t.Fatalf("got error %q, want %q", got, want)
		}
	}
}

var badYangInput = `
// Base test yang module.
// This module is syntactally correct (we can build an AST) but it is has
// invalid parameters in many statements.
module base {
  namespace "urn:mod";
  prefix "base";

  container c {
    // bad config value in a container
    config bad;
  }
  container d {
    leaf bob {
      // bad config value
      config incorrect;
      type unknown;
    }
    // duplicate leaf entry bob
    leaf bob { type string; }
    // unknown grouping to uses
    uses the-beatles;
  }
  // augmentation of unknown element
  augment nothing {
    leaf bob {
      type string;
      // bad config value in unused augment
      config wrong;
    }
  }
  grouping the-group {
    leaf one { type string; }
    // duplicate leaf in unused grouping.
    leaf one { type int; }
  }
  uses the-group;
}
`

var badYangErrors = []string{
	`bad.yang:9:3: invalid config value: bad`,
	`bad.yang:14:5: invalid config value: incorrect`,
	`bad.yang:17:7: unknown type: base:unknown`,
	`bad.yang:20:5: duplicate key: bob`,
	`bad.yang:22:5: unknown group: the-beatles`,
	`bad.yang:25:3: augment element not found: nothing`,
	`bad.yang:35:5: duplicate key: one`,
}

func TestBadYang(t *testing.T) {
	typeDict = typeDictionary{dict: map[Node]map[string]*Typedef{}}
	ms := NewModules()
	if err := ms.Parse(badYangInput, "bad.yang"); err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	errs := ms.Process()
	if len(errs) != len(badYangErrors) {
		t.Errorf("got %d errors, want %d", len(errs), len(badYangErrors))
	} else {
		ok := true
		for x, err := range errs {
			if err.Error() != badYangErrors[x] {
				ok = false
				break
			}
		}
		if ok {
			return
		}
	}

	var b bytes.Buffer
	fmt.Fprint(&b, "got errors:\n")
	for _, err := range errs {
		fmt.Fprintf(&b, "\t%v\n", err)
	}
	fmt.Fprint(&b, "want errors:\n")
	for _, err := range badYangErrors {
		fmt.Fprintf(&b, "\t%s\n", err)
	}
	t.Error(b.String())
}