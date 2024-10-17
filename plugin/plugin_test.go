// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import "testing"

func TestPlugin(t *testing.T) {
	t.Skip()
}

func TestMatch(t *testing.T) {
	tests := []struct {
		exclude []string
		include []string
		file    string
		want    bool
	}{
		{
			exclude: []string{
				".drone.yml",
			},
			include: []string{
				"*.yml",
			},
			file: "example.yml",
			want: true,
		},
		{
			exclude: []string{
				".drone.yml",
			},
			include: []string{
				"*.yml",
			},
			file: ".drone.yml",
			want: false,
		},
	}

	for _, tc := range tests {
		got := match(tc.exclude, tc.include, tc.file)

		if tc.want != got {
			t.Errorf("Want %v got %v", tc.want, got)
		}
	}
}
