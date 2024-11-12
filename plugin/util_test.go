// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import "testing"

// compareSlices checks if s2 is a subset of s1
func compareSlices(s1, s2 []string) bool {
    // Special case: if both slices are empty, they're equal
    if len(s1) == 0 && len(s2) == 0 {
        return true
    }
    
    // If s2 is empty but s1 isn't, or s1 is shorter than s2, they can't match
    if len(s2) == 0 || len(s1) < len(s2) {
        return false
    }

    // For each possible starting position in s1
    for i := 0; i <= len(s1)-len(s2); i++ {
        allMatch := true
        // Try to match all elements of s2 starting at position i
        for j := 0; j < len(s2); j++ {
            if s1[i+j] != s2[j] {
                allMatch = false
                break
            }
        }
        if allMatch {
            return true
        }
    }
    return false
}

func TestExtractIssues(t *testing.T) {
    tests := []struct {
        name string
        text string
        want []string
    }{
        {
            name: "Single issue",
            text: "TEST-1 this is a test",
            want: []string{"TEST-1"},
        },
        {
            name: "Two issues in brackets",
            text: "suffix [TEST-123] [TEST-234]",
            want: []string{"TEST-123", "TEST-234"},
        },
        {
            name: "Two issues, one in prefix",
            text: "[TEST-123] prefix [TEST-456]",
            want: []string{"TEST-123"},
        },
        {
            name: "Multiple comma-separated issues",
            text: "Multiple issues: TEST-123, TEST-234, TEST-456",
            want: []string{"TEST-123", "TEST-234", "TEST-456"},
        },
        {
            name: "Mixed format issues",
            text: "feature/TEST-123 [TEST-456] and [TEST-789]",
            want: []string{"TEST-123", "TEST-456", "TEST-789"},
        },
        {
            name: "Space-separated issues",
            text: "TEST-123 TEST-456 TEST-789",
            want: []string{"TEST-123", "TEST-456", "TEST-789"},
        },
        {
            name: "No issues",
            text: "no issue",
            want: []string{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var args Args
            args.Commit.Message = tt.text
            args.Project = "TEST"
            
            got := extractIssues(args)
            
            if !compareSlices(got, tt.want) {
                t.Errorf("\ngot:  %v\nwant: %v", got, tt.want)
            }
        })
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
			expectedOutput: "production",  // Updated to match the default value of "production"
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
			expectedOutput: "production",  // Updated to match the default value of "production"
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