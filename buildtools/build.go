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
	"fmt"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sbom"

	"github.com/paketo-buildpacks/build-tools/gradle"
	"github.com/paketo-buildpacks/build-tools/libbs"
	"github.com/paketo-buildpacks/build-tools/maven"
	"github.com/paketo-buildpacks/build-tools/sbt"
)

// ApplicationFactory builds the libbs application that the selected build tool
// drives. It is an interface so tests can substitute a stub.
type ApplicationFactory interface {
	NewApplication(additionalMetadata map[string]interface{}, arguments []string, artifactResolver libbs.ArtifactResolver,
		cache libbs.Cache, command string, bom *libcnb.BOM, applicationPath string, bomScanner sbom.SBOMScanner) (libbs.Application, error) //nolint:staticcheck // hold off on the BOM migration for now
}

// Build dispatches to the build tool selected by the detect phase.
//
// The GradleBuilder, MavenBuilder and SBTBuilder fields are primarily for
// tests; when nil, the real tool builds are used.
type Build struct {
	Logger                bard.Logger
	ApplicationFactory    ApplicationFactory
	HomeDirectoryResolver gradle.HomeDirectoryResolver
	TTY                   bool

	GradleBuilder libcnb.Builder
	MavenBuilder  libcnb.Builder
	SBTBuilder    libcnb.Builder
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	if _, found, err := pr.Resolve(maven.PlanEntryMaven); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve maven plan entry\n%w", err)
	} else if found {
		return b.mavenBuilder().Build(context)
	}

	if _, found, err := pr.Resolve(gradle.PlanEntryGradle); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve gradle plan entry\n%w", err)
	} else if found {
		return b.gradleBuilder().Build(context)
	}

	if _, found, err := pr.Resolve(sbt.PlanEntrySBT); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve sbt plan entry\n%w", err)
	} else if found {
		return b.sbtBuilder().Build(context)
	}

	return libcnb.BuildResult{}, fmt.Errorf("no supported build tool plan entry found, expected one of: %s, %s, %s",
		gradle.PlanEntryGradle, maven.PlanEntryMaven, sbt.PlanEntrySBT)
}

func (b Build) gradleBuilder() libcnb.Builder {
	if b.GradleBuilder != nil {
		return b.GradleBuilder
	}

	return gradle.Build{
		Logger:                b.Logger,
		ApplicationFactory:    b.ApplicationFactory,
		HomeDirectoryResolver: b.HomeDirectoryResolver,
	}
}

func (b Build) mavenBuilder() libcnb.Builder {
	if b.MavenBuilder != nil {
		return b.MavenBuilder
	}

	return maven.Build{
		Logger:             b.Logger,
		ApplicationFactory: b.ApplicationFactory,
		TTY:                b.TTY,
	}
}

func (b Build) sbtBuilder() libcnb.Builder {
	if b.SBTBuilder != nil {
		return b.SBTBuilder
	}

	return sbt.Build{
		Logger:             b.Logger,
		ApplicationFactory: b.ApplicationFactory,
	}
}
