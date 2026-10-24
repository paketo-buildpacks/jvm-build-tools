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

package buildtools

import (
	"github.com/buildpacks/libcnb"

	"github.com/paketo-buildpacks/build-tools/gradle"
	"github.com/paketo-buildpacks/build-tools/maven"
	"github.com/paketo-buildpacks/build-tools/sbt"
)

// Detect runs each build tool detector in order and returns the first result
// that passes. The order defines precedence when more than one build tool is
// present in the application directory.
type Detect struct {
	Detectors []libcnb.Detector
}

// NewDetect returns a Detect that checks for Gradle, Maven, then sbt.
func NewDetect() Detect {
	return Detect{
		Detectors: []libcnb.Detector{
			gradle.Detect{},
			maven.Detect{},
			sbt.Detect{},
		},
	}
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	detectors := d.Detectors
	if len(detectors) == 0 {
		detectors = NewDetect().Detectors
	}

	for _, detector := range detectors {
		result, err := detector.Detect(context)
		if err != nil {
			return libcnb.DetectResult{}, err
		}

		if result.Pass {
			return result, nil
		}
	}

	return libcnb.DetectResult{Pass: false}, nil
}
