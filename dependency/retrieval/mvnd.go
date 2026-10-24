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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	mvndID      = "mvnd"
	mvndName    = "Apache Maven Daemon"
	mvndPURL    = "pkg:generic/apache-mvnd"
	mvndLicense = "https://www.apache.org/licenses/"

	mvndOrg  = "apache"
	mvndRepo = "maven-mvnd"
)

// mvndAssets maps a buildpack architecture to the upstream asset pattern.
var mvndAssets = map[string]*regexp.Regexp{
	"amd64": regexp.MustCompile(`^maven-mvnd-.+-linux-amd64\.zip$`),
	"arm64": regexp.MustCompile(`^maven-mvnd-.+-linux-aarch64\.zip$`),
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

type mvndVersion struct {
	version *semver.Version
	tag     string
	assets  []githubAsset
}

func (v mvndVersion) Version() *semver.Version {
	return v.version
}

func getAllMvndVersions() (versionology.VersionFetcherArray, error) {
	releases, err := fetchReleases(mvndOrg, mvndRepo)
	if err != nil {
		return nil, err
	}

	var versions versionology.VersionFetcherArray
	for _, release := range releases {
		v, err := semver.NewVersion(strings.TrimPrefix(release.TagName, "v"))
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", release.TagName)
			continue
		}

		versions = append(versions, mvndVersion{version: v, tag: release.TagName, assets: release.Assets})
	}

	return versions, nil
}

func generateMvndMetadata(versionFetcher versionology.VersionFetcher, platform retrieve.Platform) ([]versionology.Dependency, error) {
	v, ok := versionFetcher.(mvndVersion)
	if !ok {
		return nil, fmt.Errorf("unexpected version type %T", versionFetcher)
	}

	versionString := v.version.String()

	asset := findAsset(v.assets, mvndAssets[platform.Arch])
	if asset == nil {
		fmt.Printf("Skipping %s: missing %s/%s asset\n", versionString, platform.OS, platform.Arch)
		return nil, nil
	}

	checksum, err := upstream.GetSHA256OfRemoteFile(asset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", asset.BrowserDownloadURL, err)
	}

	source := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", mvndOrg, mvndRepo, v.tag)
	sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Arch:     platform.Arch,
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:apache:mvnd:%s:*:*:*:*:*:*:*", versionString),
		ID:       mvndID,
		Licenses: []interface{}{
			map[string]string{
				"type": "Apache-2.0",
				"uri":  mvndLicense,
			},
		},
		Name:           mvndName,
		OS:             platform.OS,
		PURL:           fmt.Sprintf("%s@%s?arch=%s", mvndPURL, versionString, platform.Arch),
		SHA256:         checksum,
		Source:         source,
		SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
		Stacks:         []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny", "*"},
		URI:            asset.BrowserDownloadURL,
		Version:        versionString,
	}

	d, err := versionology.NewDependency(dependency, fmt.Sprintf("%s-%s", platform.OS, platform.Arch))
	if err != nil {
		return nil, err
	}

	return []versionology.Dependency{d}, nil
}

func findAsset(assets []githubAsset, pattern *regexp.Regexp) *githubAsset {
	if pattern == nil {
		return nil
	}

	for i := range assets {
		if pattern.MatchString(assets[i].Name) {
			return &assets[i]
		}
	}

	return nil
}

func fetchReleases(owner, repo string) ([]githubRelease, error) {
	var all []githubRelease
	for page := 1; ; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?page=%d&per_page=100", owner, repo, page)
		body, err := fetchBody(url)
		if err != nil {
			return nil, err
		}

		var releases []githubRelease
		if err := json.Unmarshal(body, &releases); err != nil {
			return nil, fmt.Errorf("unable to parse %s\n%w", url, err)
		}

		if len(releases) == 0 {
			break
		}

		for _, release := range releases {
			if release.Prerelease {
				continue
			}
			all = append(all, release)
		}
	}

	return all, nil
}
