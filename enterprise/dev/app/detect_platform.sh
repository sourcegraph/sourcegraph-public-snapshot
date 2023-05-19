#!/usr/bin/env bash
set -eu

detect_platform() {
  # We need to determine the platform string for the sourcegraph-backend binary
  local arch=""
  local platform=""
  case "$(uname -m)" in
    "amd64")
      arch="x86_64"
      ;;
    "arm64")
      arch="aarch64"
      ;;
    "x86_64")
      arch="x86_64"
      ;;
    *)
      echo "unknown arch: $(uname -m)"
      exit 1
  esac

  case "$(uname -s)" in
    "Darwin")
      platform="${arch}-apple-darwin"
      ;;
    "Linux")
      platform="${arch}-unknown-linux-gnu"
      ;;
    *)
      # if we get here something is wrong
      echo "unknown OS: $(uname -s)"
      exit 1
  esac

  if [[ -n ${PLATFORM_OVERRIDE:-""} ]]; then
    platform="${PLATFORM_OVERRIDE}"
  fi

  echo "${platform}"
}

detect_platform
