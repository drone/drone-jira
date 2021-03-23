// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import "testing"

func TestExtractIssue(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{
			text: "TEST-1 this is a test",
			want: "TEST-1",
		},
		{
			text: "suffix [TEST-123]",
			want: "TEST-123",
		},
		{
			text: "[TEST-123] prefix",
			want: "TEST-123",
		},
		{
			text: "TEST-123 prefix",
			want: "TEST-123",
		},
		{
			text: "feature/TEST-123",
			want: "TEST-123",
		},
		{
			text: "no issue",
			want: "",
		},
	}
	for _, test := range tests {
		var args Args
		args.Commit.Message = test.text
		args.Project = "TEST"
		if got, want := extractIssue(args), test.want; got != want {
			t.Errorf("Got issue number %v, want %v", got, want)
		}
	}
}
