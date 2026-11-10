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
	"fmt"
	"testing"

	"github.com/buildpacks/libcnb"

	"github.com/paketo-buildpacks/jvm-build-tools/buildtools"
)

type fakeDetector struct {
	result libcnb.DetectResult
	err    error
	called *bool
}

func (f fakeDetector) Detect(libcnb.DetectContext) (libcnb.DetectResult, error) {
	if f.called != nil {
		*f.called = true
	}
	return f.result, f.err
}

func TestUnitDetectReturnsFirstPassingDetector(t *testing.T) {
	secondCalled := false

	d := buildtools.Detect{
		Detectors: []libcnb.Detector{
			fakeDetector{result: libcnb.DetectResult{Pass: false}},
			fakeDetector{
				result: libcnb.DetectResult{
					Pass:  true,
					Plans: []libcnb.BuildPlan{{Provides: []libcnb.BuildPlanProvide{{Name: "maven"}}}},
				},
				called: &secondCalled,
			},
			fakeDetector{result: libcnb.DetectResult{Pass: true}},
		},
	}

	result, err := d.Detect(libcnb.DetectContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Pass {
		t.Fatal("expected detection to pass")
	}
	if !secondCalled {
		t.Fatal("expected the second detector to run")
	}
	if len(result.Plans) != 1 || result.Plans[0].Provides[0].Name != "maven" {
		t.Fatalf("unexpected plans: %#v", result.Plans)
	}
}

func TestUnitDetectReturnsFalseWhenNoDetectorPasses(t *testing.T) {
	d := buildtools.Detect{
		Detectors: []libcnb.Detector{
			fakeDetector{result: libcnb.DetectResult{Pass: false}},
			fakeDetector{result: libcnb.DetectResult{Pass: false}},
		},
	}

	result, err := d.Detect(libcnb.DetectContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pass {
		t.Fatal("expected detection to fail")
	}
}

func TestUnitDetectPropagatesErrors(t *testing.T) {
	d := buildtools.Detect{
		Detectors: []libcnb.Detector{
			fakeDetector{err: fmt.Errorf("boom")},
		},
	}

	if _, err := d.Detect(libcnb.DetectContext{}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestUnitDetectorsForToolsPreservesOrder(t *testing.T) {
	detectors, err := buildtools.DetectorsForTools("leiningen, clojure-tools, gradle, maven, sbt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(detectors) != 5 {
		t.Fatalf("expected 5 detectors, got %d", len(detectors))
	}
}

func TestUnitSupportedToolsMatchesJavaOrder(t *testing.T) {
	expected := []string{"leiningen", "clojure-tools", "gradle", "maven", "sbt"}
	actual := buildtools.SupportedTools()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d tools, got %d", len(expected), len(actual))
	}

	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf("expected tool %d to be %q, got %q", i, expected[i], actual[i])
		}
	}
}

func TestUnitDetectorsForToolsRejectsUnknownTools(t *testing.T) {
	if _, err := buildtools.DetectorsForTools("maven,nope"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestUnitDetectorsForToolsRejectsEmptyLists(t *testing.T) {
	if _, err := buildtools.DetectorsForTools(" , "); err == nil {
		t.Fatal("expected an error")
	}
}
