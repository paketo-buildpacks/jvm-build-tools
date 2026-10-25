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
	"os"

	"github.com/paketo-buildpacks/libdependency/buildpack_config"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/versionology"
)

type tool struct {
	id                   string
	getAll               retrieve.GetAllVersionsFunc
	generate             retrieve.GenerateMetadataFunc
	generateWithPlatform retrieve.GenerateMetadataWithPlatformFunc
}

var tools = []tool{
	{
		id:       "lein",
		getAll:   getAllLeiningenVersions,
		generate: generateLeiningenMetadata,
	},
	{
		id:       "clojure",
		getAll:   getAllClojureVersions,
		generate: generateClojureMetadata,
	},
	{
		id:       "gradle",
		getAll:   getAllGradleVersions,
		generate: generateGradleMetadata,
	},
	{
		id:       "maven",
		getAll:   getAllMavenVersions,
		generate: generateMavenMetadata,
	},
	{
		id:                   "mvnd",
		getAll:               getAllMvndVersions,
		generateWithPlatform: generateMvndMetadata,
	},
	{
		id:       "sbt",
		getAll:   getAllSbtVersions,
		generate: generateSbtMetadata,
	},
}

func main() {
	buildpackTomlPath, output := retrieve.FetchArgs()

	config, err := buildpack_config.ParseBuildpackToml(buildpackTomlPath)
	if err != nil {
		panic(err)
	}

	var metadata []versionology.Dependency
	for _, t := range tools {
		newVersions, err := retrieve.GetNewVersionsForId(t.id, config, t.getAll)
		if err != nil {
			panic(fmt.Errorf("unable to retrieve versions for %s\n%w", t.id, err))
		}

		if t.generateWithPlatform != nil {
			targets := config.Targets
			if len(targets) == 0 {
				targets = defaultTargets()
			}

			for _, target := range targets {
				platform := retrieve.Platform{OS: target.OS, Arch: target.Arch}
				metadata = append(metadata, retrieve.GenerateAllMetadataWithPlatform(newVersions, t.generateWithPlatform, platform)...)
			}
			continue
		}

		metadata = append(metadata, retrieve.GenerateAllMetadata(newVersions, t.generate)...)
	}

	content, err := json.Marshal(metadata)
	if err != nil {
		panic(fmt.Errorf("unable to marshal metadata\n%w", err))
	}

	if err := os.WriteFile(output, content, 0o644); err != nil {
		panic(fmt.Errorf("cannot write to %s: %w", output, err))
	}
	fmt.Printf("Wrote metadata to %s\n", output)
}
