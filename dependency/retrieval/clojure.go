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
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	clojureID   = "clojure"
	clojureName = "Clojure"

	clojureOrg  = "clojure"
	clojureRepo = "brew-install"
)

var clojureVersionPattern = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

type clojureVersion struct {
	version *semver.Version
	branch  string
}

func (v clojureVersion) Version() *semver.Version {
	return v.version
}

type clojureBranch struct {
	Name string `json:"name"`
}

func getAllClojureVersions() (versionology.VersionFetcherArray, error) {
	branches, err := getAllClojureBranches()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch branches\n%w", err)
	}

	var versions versionology.VersionFetcherArray
	for _, branch := range branches {
		if !clojureVersionPattern.MatchString(branch) {
			continue
		}

		v, err := semver.NewVersion(branch)
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", branch)
			continue
		}

		versions = append(versions, clojureVersion{version: v, branch: branch})
	}

	return versions, nil
}

func generateClojureMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	version, ok := versionFetcher.(clojureVersion)
	if !ok {
		return nil, fmt.Errorf("unexpected version type %T", versionFetcher)
	}

	versionString := version.version.String()

	installerVersion, err := fetchClojureStableVersion(version.branch)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch stable version for %s\n%w", versionString, err)
	}

	if !strings.HasPrefix(installerVersion, versionString+".") {
		fmt.Printf("Skipping %s: stable installer version %s does not match\n", versionString, installerVersion)
		return nil, nil
	}

	uri := fmt.Sprintf("https://download.clojure.org/install/linux-install-%s.sh", installerVersion)
	checksum, err := upstream.GetSHA256OfRemoteFile(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", uri, err)
	}

	source := fmt.Sprintf("https://github.com/clojure/clojure/archive/refs/tags/clojure-%s.tar.gz", version.branch)
	sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:cognitect:clojure:%s:*:*:*:*:*:*:*", versionString),
		ID:       clojureID,
		Licenses: []interface{}{
			map[string]string{
				"type": "Eclipse Public License - v 1.0",
				"uri":  "https://github.com/clojure/clojure/blob/master/epl-v10.html",
			},
		},
		Name:           clojureName,
		PURL:           retrieve.GeneratePURL(clojureID, versionString, checksum, uri),
		SHA256:         checksum,
		Source:         source,
		SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
		Stacks:         []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny", "*"},
		URI:            uri,
		Version:        versionString,
	}

	return versionology.NewDependencyArray(dependency, "")
}

func getAllClojureBranches() ([]string, error) {
	var branches []string
	for page := 1; ; page++ {
		var pageBranches []clojureBranch
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches?page=%d&per_page=100", clojureOrg, clojureRepo, page)
		if err := fetchJSON(url, &pageBranches); err != nil {
			return nil, err
		}

		if len(pageBranches) == 0 {
			break
		}

		for _, b := range pageBranches {
			branches = append(branches, b.Name)
		}
	}

	return branches, nil
}

func fetchClojureStableVersion(branch string) (string, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/stable.properties", clojureOrg, clojureRepo, branch)
	b, err := fetchBody(url)
	if err != nil {
		return "", fmt.Errorf("unable to download %s: %w", url, err)
	}

	parts := strings.Fields(string(b))
	if len(parts) == 0 {
		return "", fmt.Errorf("unable to parse stable version from %s", url)
	}

	return parts[0], nil
}
