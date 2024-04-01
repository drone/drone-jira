// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import "testing"

func compareSlices(s1, s2 []string) bool {
    if len(s2) < 1 || len(s1) < len(s2){
        return false
    }
    
    for i := 0; i < len(s1); i++ {
        found := true
        for j := 0; j < len(s2); j++ {
            if s1[i+j] != s2[j] {
                found = false
                break
            }
        }
        if found {
            return true
        }
    }
    return false
}

func TestExtractIssues(t *testing.T) {
    tests := []struct {
        text string
        want []string
    }{
        {
            text: "TEST-1 this is a test",
            want: []string{"TEST-1"},
        },
        {
            text: "suffix [TEST-123] [TEST-234]",
            want: []string{"TEST-123", "TEST-234"},
        },
        {
            text: "[TEST-123] prefix [TEST-456]",
            want: []string{"TEST-123"},
        },
        {
            text: "Multiple issues: TEST-123, TEST-234, TEST-456",
            want: []string{"TEST-123"},
        },
        {
            text: "feature/TEST-123 [TEST-456] and [TEST-789]",
            want: []string{"TEST-123"},
        },
        {
            text: "TEST-123 TEST-456 TEST-789",
            want: []string{"TEST-123"},
        },
        {
            text: "no issue",
            want: []string{},
        },
    }

    t.Errorf("TESTING")

    for _, test := range tests {
        var args Args
        args.Commit.Message = test.text
        args.Project = "TEST"
        got := extractIssues(args)
        t.Errorf("TEXT:%v || WANT: %s", got, test.want) 
        t.Errorf(" %v", compareSlices(got, test.want))
        if !compareSlices(got, test.want) {
            t.Errorf("Got issues %v, want %v", got, test.want)
        }
    }

}
