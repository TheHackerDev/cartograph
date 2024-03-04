/*******************************************************************************
 * Copyright 2018-2024 Aaron Hnatiw
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 ******************************************************************************/

package datatypes

import (
	"reflect"
	"testing"
)

func TestSplitParameterString(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "foo-bar",
			want:  []string{"foo", "bar"},
		},
		{
			input: "foo_bar",
			want:  []string{"foo", "bar"},
		},
		{
			input: "foobar",
			want:  []string{"foobar"},
		},
		{
			input: "foo-bar_1-2-3_abcdef",
			want:  []string{"foo", "bar", "abcdef"},
		},
		{
			input: "areallylongparameterstringthatismorethanfiftycharacters",
			want:  []string{},
		},
		{
			input: "123-456",
			want:  []string{},
		},
	}

	for _, test := range tests {
		got := splitParameterString(test.input)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("splitParameterString(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}

func TestSplitPathString(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "/foo/bar",
			want:  []string{"foo", "bar"},
		},
		{
			input: "/foo_bar",
			want:  []string{"foo", "bar"},
		},
		{
			input: "/foobar",
			want:  []string{"foobar"},
		},
		{
			input: "/foo-bar_1-2-3_abcdef",
			want:  []string{"foo", "bar", "abcdef"},
		},
		{
			input: "/areallylongpathsegmentthatismorethanfiftycharacters",
			want:  []string{},
		},
		{
			input: "/123-456",
			want:  []string{},
		},
		{
			input: "",
			want:  []string{},
		},
		{
			input: "/",
			want:  []string{},
		},
	}

	for _, test := range tests {
		got := splitPathString(test.input)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("splitPathString(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}
