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
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	sbtID      = "sbt"
	sbtName    = "Scala SBT"
	sbtPURL    = "pkg:generic/sbt"
	sbtLicense = "https://github.com/sbt/sbt/blob/develop/LICENSE"

	sbtOrg  = "sbt"
	sbtRepo = "sbt"
)

var (
	sbtVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	sbtAssetPattern   = regexp.MustCompile(`^sbt-\d+\.\d+\.\d+\.tgz$`)
)

type sbtVersion struct {
	version *semver.Version
	tag     string
	uri     string
}

func (v sbtVersion) Version() *semver.Version {
	return v.version
}

func getAllSbtVersions() (versionology.VersionFetcherArray, error) {
	releases, err := fetchReleases(sbtOrg, sbtRepo)
	if err != nil {
		return nil, err
	}

	var versions versionology.VersionFetcherArray
	for _, release := range releases {
		tag := strings.TrimPrefix(release.TagName, "v")
		if !sbtVersionPattern.MatchString(tag) {
			continue
		}

		v, err := semver.NewVersion(tag)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", release.TagName)
			continue
		}

		asset := findAsset(release.Assets, sbtAssetPattern)
		if asset == nil {
			fmt.Printf("Skipping %s: no matching asset\n", release.TagName)
			continue
		}

		versions = append(versions, sbtVersion{version: v, tag: tag, uri: asset.BrowserDownloadURL})
	}

	return versions, nil
}

func generateSbtMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	v, ok := versionFetcher.(sbtVersion)
	if !ok {
		return nil, fmt.Errorf("unexpected version type %T", versionFetcher)
	}

	versionString := v.version.String()

	checksum, err := upstream.GetSHA256OfRemoteFile(v.uri)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", v.uri, err)
	}

	source := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/v%s.tar.gz", sbtOrg, sbtRepo, v.tag)
	sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:lightbend:sbt:%s:*:*:*:*:*:*:*", versionString),
		ID:       sbtID,
		Licenses: []interface{}{
			map[string]string{
				"type": "Apache-2.0",
				"uri":  sbtLicense,
			},
		},
		Name:           sbtName,
		PURL:           fmt.Sprintf("%s@%s", sbtPURL, versionString),
		SHA256:         checksum,
		Source:         source,
		SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
		Stacks:         []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny", "*"},
		URI:            v.uri,
		Version:        versionString,
	}

	return versionology.NewDependencyArray(dependency, "")
}
