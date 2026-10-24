# `github.com/paketo-buildpacks/build-tools`

The Paketo Buildpack for Build Tools is a Cloud Native Buildpack that provides
build tools (Gradle, Maven, Maven Daemon and sbt) and builds JVM applications
with them.

This buildpack is designed to work in collaboration with other buildpacks which
request contributions of build tools, and replaces the previously separate
`gradle`, `maven` and `sbt` buildpacks. The detection, build and distribution
code is shared from `libbs`; the per-tool behavior lives in the `gradle`,
`maven` and `sbt` packages and all of them are served by a single buildpack
binary.

## Behavior

The buildpack will participate if any of the following conditions are met:

- a Gradle build file (`build.gradle`, `build.gradle.kts`, `settings.gradle` or
  `settings.gradle.kts`, or `$BP_GRADLE_BUILD_FILE`) is present
- a Maven `pom.xml` (or `$BP_MAVEN_POM_FILE`) is present
- an `sbt` launcher, `build.sbt`, or sbt project file is present

Detection precedence is Gradle, then Maven, then sbt.

For the detected tool the buildpack will:

- contribute the requested build tool (Gradle distribution, Apache Maven, Maven
  Daemon or sbt launcher) to a layer on `$PATH`, unless the application ships
  its own wrapper/launcher
- run the tool to build the application, honoring the tool's
  `BP_<TOOL>_BUILD_ARGUMENTS` and related configuration
- contribute the built application artifact and cache the compiled application
  layer, along with the SBOM entries for the build tool dependency

## Configuration

| Environment Variable                | Description                                                                                  |
| ----------------------------------- | -------------------------------------------------------------------------------------------- |
| `$BP_GRADLE_BUILD_ARGUMENTS`        | the arguments to pass to Gradle                                                              |
| `$BP_GRADLE_ADDITIONAL_BUILD_ARGUMENTS` | additional arguments appended to `$BP_GRADLE_BUILD_ARGUMENTS`                            |
| `$BP_GRADLE_BUILD_FILE`             | the location of the main build config file, relative to the application root                  |
| `$BP_GRADLE_BUILT_ARTIFACT`         | the built application artifact explicitly                                                    |
| `$BP_GRADLE_BUILT_MODULE`           | the module to find the application artifact in                                               |
| `$BP_GRADLE_INIT_SCRIPT_PATH`       | the path to a Gradle init script file                                                        |
| `$BP_MAVEN_BUILD_ARGUMENTS`         | the arguments to pass to Maven                                                               |
| `$BP_MAVEN_ADDITIONAL_BUILD_ARGUMENTS` | additional arguments appended to `$BP_MAVEN_BUILD_ARGUMENTS`                              |
| `$BP_MAVEN_ACTIVE_PROFILES`         | the active Maven profiles (comma separated)                                                  |
| `$BP_MAVEN_BUILT_ARTIFACT`          | the built application artifact explicitly                                                    |
| `$BP_MAVEN_BUILT_MODULE`            | the module to find the application artifact in                                               |
| `$BP_MAVEN_DAEMON_ENABLED`          | use the Maven Daemon (`mvnd`) instead of `mvn`                                               |
| `$BP_MAVEN_POM_FILE`                | the location of the main `pom.xml`, relative to the application root                         |
| `$BP_MAVEN_SETTINGS_PATH`           | the path to a Maven settings file                                                            |
| `$BP_MAVEN_VERSION`                 | the Maven version line to install (defaults to `3`)                                          |
| `$BP_SBT_BUILD_ARGUMENTS`           | the arguments to pass to sbt                                                                 |
| `$BP_SBT_BUILT_ARTIFACT`            | the built application artifact explicitly                                                    |
| `$BP_SBT_BUILT_MODULE`              | the module to find the application artifact in                                               |
| `$BP_SBT_REPOSITORIES_FILE`         | the path to the sbt repositories file                                                        |
| `$BP_SBT_VERSION`                   | the sbt version line to install (defaults to `1`)                                            |
| `$BP_INCLUDE_FILES`                 | colon separated list of glob patterns for source files to include                            |
| `$BP_EXCLUDE_FILES`                 | colon separated list of glob patterns for source files to remove                             |
| `$BP_JAVA_INSTALL_NODE`             | whether to request Yarn/Node based on `package.json` or `yarn.lock`                          |
| `$BP_NODE_PROJECT_PATH`             | project subdirectory to look for `package.json` and `yarn.lock`                              |

## Layout

```
.
├── buildtools/        combined detect/build dispatch for all tools
├── gradle/            Gradle detection, distribution and build
├── libbs/             shared application, cache and resolver code (from libbs)
├── maven/             Maven detection, distributions (mvn/mvnd) and build
├── sbt/               sbt detection, distribution and build
├── dependency/        dependency retrieval for gradle, maven, mvnd and sbt
└── run/               buildpack entrypoint
```

## Dependencies

Dependencies are retrieved with `make retrieve` from `dependency/` and written to
`buildpack.toml` with `jam update-dependencies`. The `gradle`, `maven`, `mvnd`
and `sbt` dependency IDs are maintained by `[[metadata.dependency-constraints]]`.
