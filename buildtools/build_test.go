/*
 * Copyright 2018-2026 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package buildtools_test

import (
	"testing"

	"github.com/buildpacks/libcnb"

	"github.com/paketo-buildpacks/jvm-build-tools/buildtools"
)

type fakeBuilder struct {
	called bool
	layer  string
}

func (f *fakeBuilder) Build(libcnb.BuildContext) (libcnb.BuildResult, error) {
	f.called = true
	return libcnb.BuildResult{Labels: []libcnb.Label{{Key: "tool", Value: f.layer}}}, nil
}

func TestUnitBuildDispatchesByPlanEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    string
		expected string
	}{
		{name: "leiningen", entry: "lein", expected: "leiningen"},
		{name: "clojure-tools", entry: "clojure", expected: "clojure-tools"},
		{name: "gradle", entry: "gradle", expected: "gradle"},
		{name: "maven", entry: "maven", expected: "maven"},
		{name: "sbt", entry: "sbt", expected: "sbt"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			leiningenBuilder := &fakeBuilder{layer: "leiningen"}
			clojureBuilder := &fakeBuilder{layer: "clojure-tools"}
			gradleBuilder := &fakeBuilder{layer: "gradle"}
			mavenBuilder := &fakeBuilder{layer: "maven"}
			sbtBuilder := &fakeBuilder{layer: "sbt"}

			b := buildtools.Build{
				LeiningenBuilder: leiningenBuilder,
				ClojureBuilder:   clojureBuilder,
				GradleBuilder:    gradleBuilder,
				MavenBuilder:     mavenBuilder,
				SBTBuilder:       sbtBuilder,
			}

			context := libcnb.BuildContext{
				Plan: libcnb.BuildpackPlan{
					Entries: []libcnb.BuildpackPlanEntry{{Name: tc.entry}},
				},
			}

			result, err := b.Build(context)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Labels) != 1 || result.Labels[0].Value != tc.expected {
				t.Fatalf("expected %s builder to be used, got %#v", tc.expected, result.Labels)
			}

			used := map[string]bool{
				"leiningen":     leiningenBuilder.called,
				"clojure-tools": clojureBuilder.called,
				"gradle":        gradleBuilder.called,
				"maven":         mavenBuilder.called,
				"sbt":           sbtBuilder.called,
			}
			for name, called := range used {
				if name == tc.expected && !called {
					t.Fatalf("expected %s builder to be called", name)
				}
				if name != tc.expected && called {
					t.Fatalf("did not expect %s builder to be called", name)
				}
			}
		})
	}
}

func TestUnitBuildFailsWithoutAToolPlanEntry(t *testing.T) {
	b := buildtools.Build{
		LeiningenBuilder: &fakeBuilder{},
		ClojureBuilder:   &fakeBuilder{},
		GradleBuilder:    &fakeBuilder{},
		MavenBuilder:     &fakeBuilder{},
		SBTBuilder:       &fakeBuilder{},
	}

	if _, err := b.Build(libcnb.BuildContext{}); err == nil {
		t.Fatal("expected an error")
	}
}
