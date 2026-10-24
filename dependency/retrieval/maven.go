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
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	mavenID      = "maven"
	mavenName    = "Apache Maven"
	mavenPURL    = "pkg:generic/apache-maven"
	mavenLicense = "https://www.apache.org/licenses/"

	mavenRepository = "https://repo1.maven.org/maven2"
	mavenGroupID    = "org.apache.maven"
	mavenArtifactID = "apache-maven"
)

var (
	mavenBase        = fmt.Sprintf("%s/%s/%s", mavenRepository, strings.ReplaceAll(mavenGroupID, ".", "/"), mavenArtifactID)
	mavenVersionList = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)
)

type mavenMetadata struct {
	Versioning struct {
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

// mavenVersion keeps the 3-part dependency version while retaining the full
// upstream version, which is needed to build the download URL (Maven 4
// prereleases have historically been published as e.g. 4.0.0-rc-5).
type mavenVersion struct {
	version *semver.Version
	full    string
}

func (v mavenVersion) Version() *semver.Version {
	return v.version
}

func getAllMavenVersions() (versionology.VersionFetcherArray, error) {
	metadataURL := fmt.Sprintf("%s/maven-metadata.xml", mavenBase)

	body, err := fetchBody(metadataURL)
	if err != nil {
		return nil, err
	}

	var metadata mavenMetadata
	if err := xml.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("unable to parse %s\n%w", metadataURL, err)
	}

	// Group by major.minor.patch and keep the newest full version, so a final
	// release wins over its prereleases and the newest prerelease wins when no
	// final release exists yet.
	latest := map[string]mavenVersion{}
	for _, full := range metadata.Versioning.Versions.Version {
		m := mavenVersionList.FindStringSubmatch(full)
		if m == nil {
			continue
		}

		versionString := fmt.Sprintf("%s.%s.%s", m[1], m[2], m[3])
		version, err := semver.NewVersion(versionString)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", full)
			continue
		}

		candidate := mavenVersion{version: version, full: full}
		existing, ok := latest[versionString]
		if !ok {
			latest[versionString] = candidate
			continue
		}

		existingFull, err := semver.NewVersion(existing.full)
		if err != nil {
			continue
		}
		candidateFull, err := semver.NewVersion(candidate.full)
		if err != nil {
			continue
		}
		if candidateFull.GreaterThan(existingFull) {
			latest[versionString] = candidate
		}
	}

	var versions versionology.VersionFetcherArray
	for _, v := range latest {
		versions = append(versions, v)
	}

	return versions, nil
}

func generateMavenMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	v, ok := versionFetcher.(mavenVersion)
	if !ok {
		return nil, fmt.Errorf("unexpected version type %T", versionFetcher)
	}

	versionString := v.version.String()

	uri := fmt.Sprintf("%s/%s/%s-%s-bin.tar.gz", mavenBase, v.full, mavenArtifactID, v.full)
	checksum, err := upstream.GetSHA256OfRemoteFile(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", uri, err)
	}

	source := fmt.Sprintf("%s/%s/%s-%s-src.tar.gz", mavenBase, v.full, mavenArtifactID, v.full)
	sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:apache:maven:%s:*:*:*:*:*:*:*", versionString),
		ID:       mavenID,
		Licenses: []interface{}{
			map[string]string{
				"type": "Apache-2.0",
				"uri":  mavenLicense,
			},
		},
		Name:           mavenName,
		PURL:           fmt.Sprintf("%s@%s", mavenPURL, versionString),
		SHA256:         checksum,
		Source:         source,
		SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
		Stacks:         []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny", "*"},
		URI:            uri,
		Version:        versionString,
	}

	return versionology.NewDependencyArray(dependency, "")
}
