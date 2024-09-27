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

func TestExtractInstanceName(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		// Test cases with URLs
		{"http://test.com", "test"},
		{"https://subdomain.test.com", "subdomain"},
		{"ftp://ftp.test.org", "ftp"},

		// Test cases with non-URL strings
		{"instance.test.com", "instance"},
		{"subdomain.instance.test.org", "subdomain"},
		{"localhost", "localhost"},

		// Test invalid or malformed URLs
		{"http://", ""},                // Invalid URL with no hostname
		{"invalid-url", "invalid-url"}, // Not a URL, should return the input string
	}

	for _, test := range tests {
		result := ExtractInstanceName(test.text)
		if result != test.want {
			t.Errorf("ExtractInstanceName(%q) = %q; expected %q", test.text, result, test.want)
		}
	}
}

// Test the toEnvironmentId function
func TestToEnvironmentId(t *testing.T) {
	tests := []struct {
		name           string
		args           Args
		expectedOutput string
	}{
		{
			name:           "Non-empty EnvironmentId",
			args:           Args{EnvironmentId: "env-123"},
			expectedOutput: "env-123",
		},
		{
			name:           "Empty EnvironmentId",
			args:           Args{EnvironmentId: ""},
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toEnvironmentId(tt.args)
			if result != tt.expectedOutput {
				t.Errorf("toEnvironmentId() = %v, want %v", result, tt.expectedOutput)
			}
		})
	}
}

// Test the toEnvironmentType function
func TestToEnvironmentType(t *testing.T) {
	tests := []struct {
		name           string
		args           Args
		expectedOutput string
	}{
		{
			name:           "Non-empty EnvironmentType",
			args:           Args{EnvironmentType: "prod"},
			expectedOutput: "prod",
		},
		{
			name:           "Empty EnvironmentType",
			args:           Args{EnvironmentType: ""},
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toEnvironmentType(tt.args)
			if result != tt.expectedOutput {
				t.Errorf("toEnvironmentType() = %v, want %v", result, tt.expectedOutput)
			}
		})
	}
}
