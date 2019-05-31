/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"path/filepath"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestSupportedKubernetesFormats(t *testing.T) {
	var tests = []struct {
		description string
		in          string
		out         bool
	}{
		{
			description: "yaml",
			in:          "filename.yaml",
			out:         true,
		},
		{
			description: "yml",
			in:          "filename.yml",
			out:         true,
		},
		{
			description: "json",
			in:          "filename.json",
			out:         true,
		},
		{
			description: "txt",
			in:          "filename.txt",
			out:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actual := IsSupportedKubernetesFormat(tt.in)
			if tt.out != actual {
				t.Errorf("out: %t, actual: %t", tt.out, actual)
			}
		})
	}
}

func TestExpandPathsGlob(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	tmpDir.Write("dir/sub_dir/file", "")
	tmpDir.Write("dir_b/sub_dir_b/file", "")

	var tests = []struct {
		description string
		in          []string
		out         []string
		shouldErr   bool
	}{
		{
			description: "match exact filename",
			in:          []string{"dir/sub_dir/file"},
			out:         []string{tmpDir.Path("dir/sub_dir/file")},
		},
		{
			description: "match leaf directory glob",
			in:          []string{"dir/sub_dir/*"},
			out:         []string{tmpDir.Path("dir/sub_dir/file")},
		},
		{
			description: "match top level glob",
			in:          []string{"dir*"},
			out:         []string{tmpDir.Path("dir/sub_dir/file"), tmpDir.Path("dir_b/sub_dir_b/file")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actual, err := ExpandPathsGlob(tmpDir.Root(), tt.in)

			testutil.CheckErrorAndDeepEqual(t, tt.shouldErr, err, tt.out, actual)
		})
	}
}

func TestExpand(t *testing.T) {
	var tests = []struct {
		description string
		text        string
		key         string
		value       string
		expected    string
	}{
		{
			description: "${key} syntax",
			text:        "BEFORE[${key}]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[VALUE]AFTER",
		},
		{
			description: "$key syntax",
			text:        "BEFORE[$key]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[VALUE]AFTER",
		},
		{
			description: "replace all",
			text:        "BEFORE[$key][${key}][$key][${key}]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[VALUE][VALUE][VALUE][VALUE]AFTER",
		},
		{
			description: "ignore common prefix",
			text:        "BEFORE[$key1][${key1}]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[$key1][${key1}]AFTER",
		},
		{
			description: "just the ${key} placeholder",
			text:        "${key}",
			key:         "key",
			value:       "VALUE",
			expected:    "VALUE",
		},
		{
			description: "just the $key placeholder",
			text:        "$key",
			key:         "key",
			value:       "VALUE",
			expected:    "VALUE",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := Expand(test.text, test.key, test.value)

			testutil.CheckDeepEqual(t, test.expected, actual)
		})
	}
}

func TestAbsFile(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()
	tmpDir.Write("file", "")
	expectedFile, err := filepath.Abs(filepath.Join(tmpDir.Root(), "file"))
	testutil.CheckError(t, false, err)

	file, err := AbsFile(tmpDir.Root(), "file")
	testutil.CheckErrorAndDeepEqual(t, false, err, expectedFile, file)

	_, err = AbsFile(tmpDir.Root(), "")
	testutil.CheckErrorAndDeepEqual(t, true, err, tmpDir.Root()+" is a directory", err.Error())

	_, err = AbsFile(tmpDir.Root(), "does-not-exist")
	testutil.CheckError(t, true, err)
}

func TestNonEmptyLines(t *testing.T) {
	var testCases = []struct {
		in  string
		out []string
	}{
		{"", nil},
		{"a\n", []string{"a"}},
		{"a\r\n", []string{"a"}},
		{"a\r\nb", []string{"a", "b"}},
		{"a\r\nb\n\n", []string{"a", "b"}},
		{"\na\r\n\n\n", []string{"a"}},
	}
	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			result := NonEmptyLines([]byte(tt.in))
			testutil.CheckDeepEqual(t, tt.out, result)
		})
	}
}

func TestCloneThroughJSON(t *testing.T) {
	tests := []struct {
		name     string
		old      interface{}
		new      interface{}
		expected interface{}
	}{
		{
			name: "google cloud build",
			old: map[string]string{
				"projectId": "unit-test",
			},
			new: &latest.GoogleCloudBuild{},
			expected: &latest.GoogleCloudBuild{
				ProjectID: "unit-test",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CloneThroughJSON(test.old, test.new)
			testutil.CheckErrorAndDeepEqual(t, false, err, test.expected, test.new)
		})
	}
}

func TestIsHiddenDir(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "hidden dir",
			filename: ".hidden",
			expected: true,
		},
		{
			name:     "not hidden dir",
			filename: "not_hidden",
			expected: false,
		},
		{
			name:     "current dir",
			filename: ".",
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if IsHiddenDir(test.filename) != test.expected {
				t.Errorf("error want %t,  got %t", test.expected, !test.expected)
			}
		})
	}
}

func TestIsHiddenFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "hidden file name",
			filename: ".hidden",
			expected: true,
		},
		{
			name:     "not hidden file",
			filename: "not_hidden",
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if IsHiddenDir(test.filename) != test.expected {
				t.Errorf("error want %t,  got %t", test.expected, !test.expected)
			}
		})
	}
}

func TestIsWorkspaceFile(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	tmpDir.Write("file", "")

	var tests = []struct {
		description string
		in          string
		expected    bool
	}{
		{
			description: "relative path",
			in:          "file",
			expected:    true,
		},
		{
			description: "absolute path",
			in:          tmpDir.Path("file"),
			expected:    true,
		},
		{
			description: "does not exists",
			in:          "does-not-exists",
			expected:    false,
		},
		{
			description: "remote git",
			in:          "git@example.com/example-repo",
			expected:    false,
		},
		{
			description: "remote https",
			in:          "https://example.com/example-repo",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actual := IsLocalFile(tt.in, tmpDir.Root())

			testutil.CheckDeepEqual(t, tt.expected, actual)
		})
	}
}

func TestRemoveFromSlice(t *testing.T) {
	testutil.CheckDeepEqual(t, []string{""}, RemoveFromSlice([]string{""}, "ANY"))
	testutil.CheckDeepEqual(t, []string{"A", "B", "C"}, RemoveFromSlice([]string{"A", "B", "C"}, "ANY"))
	testutil.CheckDeepEqual(t, []string{"A", "C"}, RemoveFromSlice([]string{"A", "B", "C"}, "B"))
	testutil.CheckDeepEqual(t, []string{"B", "C"}, RemoveFromSlice([]string{"A", "B", "C"}, "A"))
	testutil.CheckDeepEqual(t, []string{"A", "C"}, RemoveFromSlice([]string{"A", "B", "B", "C"}, "B"))
	testutil.CheckDeepEqual(t, []string{}, RemoveFromSlice([]string{"B", "B"}, "B"))
}
