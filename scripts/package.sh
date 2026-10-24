#!/usr/bin/env bash

set -eu
set -o pipefail

readonly ROOT_DIR="$(cd "$(dirname "${0}")/.." && pwd)"
readonly BIN_DIR="${ROOT_DIR}/.bin"
readonly BUILD_DIR="${ROOT_DIR}/build"

# shellcheck source=SCRIPTDIR/.util/tools.sh
source "${ROOT_DIR}/scripts/.util/tools.sh"

# shellcheck source=SCRIPTDIR/.util/print.sh
source "${ROOT_DIR}/scripts/.util/print.sh"

function main {
  local version output token
  token=""

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --version|-v)
        version="${2}"
        shift 2
        ;;

      --output|-o)
        output="${2}"
        shift 2
        ;;

      --token|-t)
        token="${2}"
        shift 2
        ;;

      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  if [[ -z "${version:-}" ]]; then
    usage
    echo
    util::print::error "--version is required"
  fi

  if [[ -z "${output:-}" ]]; then
    output="${BUILD_DIR}/buildpackage.cnb"
  fi

  repo::prepare

  tools::install "${token}"

  buildpack_type=buildpack
  if [ -f "${ROOT_DIR}/extension.toml" ]; then
    buildpack_type=extension
  fi

  buildpackage::create::individual_tools "${output}" "${buildpack_type}" "${version}"
  buildpackage::create::single_all_tools "${output}" "${buildpack_type}" "${version}"
}

function usage() {
  cat <<-USAGE
package.sh --version <version> [OPTIONS]

Packages a buildpack or an extension into a buildpackage .cnb file.

OPTIONS
  --help               -h            prints the command usage
  --version <version>  -v <version>  specifies the version number to use when packaging a buildpack or an extension
  --output <output>    -o <output>   location to output the packaged buildpackage or extension artifact (default: ${ROOT_DIR}/build/buildpackage.cnb)
  --token <token>                    Token used to download assets from GitHub (e.g. jam, pack, etc) (optional)
USAGE
}

function repo::prepare() {
  util::print::title "Preparing repo..."

  rm -rf "${BUILD_DIR}"

  mkdir -p "${BIN_DIR}"
  mkdir -p "${BUILD_DIR}"

  export PATH="${BIN_DIR}:${PATH}"
}

function tools::install() {
  local token
  token="${1}"

  util::tools::pack::install \
    --directory "${BIN_DIR}" \
    --token "${token}"

  if [[ -f "${ROOT_DIR}/.libbuildpack" ]]; then
    util::tools::packager::install \
      --directory "${BIN_DIR}"
  else
    util::tools::libpak-tools::install \
      --directory "${BIN_DIR}" \
      --token "${token}"
  fi
}

function buildpackage::create::individual_tools() {
  local output version
  output="${1}"
  buildpack_type="${2}"
  version="${3}"

  util::print::title "Packaging ${buildpack_type}..."

  libpak-tools build-build-tools \
      --buildpack-id "paketo-buildpacks/maven@${version}" --tools maven \
      --buildpack-id "paketo-buildpacks/gradle@${version}" --tools gradle \
      --buildpack-id "paketo-buildpacks/sbt@${version}" --tools sbt \
      --format file \
      --buildpack-path . \
      --output-path "$PWD/build"
}

function buildpackage::create::single_all_tools() {
  local output version
  output="${1}"
  buildpack_type="${2}"
  version="${3}"

  util::print::title "Packaging ${buildpack_type}..."

  libpak-tools build-build-tools \
      --single-buildpack \
      --buildpack-id "paketo-buildpacks/build-tools@${version}" \
      --include-all-tools \
      --format file \
      --buildpack-path . \
      --output-path "$PWD/build"
}

main "${@:-}"
