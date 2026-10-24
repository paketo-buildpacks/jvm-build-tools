// Copyright 2018-2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	gradleID      = "gradle"
	gradleName    = "Gradle"
	gradlePURL    = "pkg:generic/gradle"
	gradleLicense = "https://docs.gradle.org/current/userguide/userguide.html#licenses"

	gradleReleasesURL = "https://raw.githubusercontent.com/gradle/gradle/master/released-versions.json"
	gradleDownloadURL = "https://downloads.gradle.org/distributions"
)

var gradleVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type gradleReleasedVersions struct {
	FinalReleases []struct {
		Version string `json:"version"`
	} `json:"finalReleases"`
}

func getAllGradleVersions() (versionology.VersionFetcherArray, error) {
	var releases gradleReleasedVersions
	if err := fetchJSON(gradleReleasesURL, &releases); err != nil {
		return nil, err
	}

	var versions versionology.VersionFetcherArray
	seen := map[string]bool{}
	for _, release := range releases.FinalReleases {
		if !gradleVersionPattern.MatchString(release.Version) || seen[release.Version] {
			continue
		}
		seen[release.Version] = true

		version, err := semver.NewVersion(release.Version)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", release.Version)
			continue
		}

		versions = append(versions, versionology.NewSimpleVersionFetcher(version))
	}

	return versions, nil
}

func generateGradleMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	versionString := versionFetcher.Version().String()

	uri := fmt.Sprintf("%s/gradle-%s-bin.zip", gradleDownloadURL, versionString)
	checksum, err := upstream.GetSHA256OfRemoteFile(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", uri, err)
	}

	source := fmt.Sprintf("%s/gradle-%s-src.zip", gradleDownloadURL, versionString)
	sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:gradle:gradle:%s:*:*:*:*:*:*:*", versionString),
		ID:       gradleID,
		Licenses: []interface{}{
			map[string]string{
				"type": "Apache-2.0",
				"uri":  gradleLicense,
			},
		},
		Name:           gradleName,
		PURL:           fmt.Sprintf("%s@%s", gradlePURL, versionString),
		SHA256:         checksum,
		Source:         source,
		SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
		Stacks:         []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny", "*"},
		URI:            uri,
		Version:        versionString,
	}

	return versionology.NewDependencyArray(dependency, "")
}
