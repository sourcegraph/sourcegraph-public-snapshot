#!/bin/sh
# shellcheck shell=dash

# bootstrap.sh downloads the latest release of `sg` into a temp location and
# runs `sg install`.

set -u

usage() {
  cat 1>&2 <<EOF
sg-install
The installer for sg

USAGE:
    sg-install
EOF
}

main() {
  need_cmd uname
  need_cmd mktemp
  need_cmd chmod
  need_cmd mkdir
  need_cmd rm
  need_cmd rmdir
  need_cmd grep
  need_cmd sed
  need_cmd curl

  get_architecture || return 1
  local _arch="$RETVAL"
  assert_nz "$_arch" "arch"

  printf '%s\n' 'determining latest release of sg' 1>&2

  local _location_header
  _location_header="$(curl --silent -I "https://github.com/sourcegraph/sg/releases/latest" | grep -i "location:" | tr -d '\r')"

  local _base_url
  _base_url="$(echo "${_location_header}" | sed s/location:\ // | sed s/tag/download/ | tr -d "[:blank:]")"
  assert_nz "$_base_url" "base_url"

  local _url
  _url="${_base_url}/sg_${_arch}"

  local _dir
  _dir="$(ensure mktemp -d)"
  local _file="${_dir}/sg"

  printf '%s\n' 'downloading sg...' 1>&2

  ensure mkdir -p "$_dir"
  ensure download "$_url" "$_file" "$_arch"
  ensure chmod u+x "$_file"
  if [ ! -x "$_file" ]; then
    printf '%s\n' "Cannot execute $_file (likely because of mounting /tmp as noexec)." 1>&2
    printf '%s\n' "Please copy the file to a location where you can execute binaries and run ./sg." 1>&2
    exit 1
  fi

  printf 'running "%s %s"\n' 'sg install' "$*" 1>&2
  "$_file" install "$*" </dev/tty
}

get_architecture() {
  local _ostype _cputype _arch
  _ostype="$(uname -s)"
  _cputype="$(uname -m)"

  case "$_ostype" in
    Darwin)
      _ostype=darwin
      ;;

    Linux)
      _ostype=linux
      ;;
    *)
      err "unrecognized or unsupported OS type: $_ostype"
      ;;

  esac

  case "$_cputype" in
    aarch64 | arm64)
      _cputype=arm64
      ;;

    x86_64 | x86-64 | x64 | amd64)
      _cputype=amd64
      ;;

    *)
      err "unknown or unsupported CPU type: $_cputype"
      ;;
  esac

  _arch="${_ostype}_${_cputype}"

  RETVAL="$_arch"
}

say() {
  printf 'sg-install: %s\n' "$1"
}

err() {
  say "$1" >&2
  exit 1
}

need_cmd() {
  if ! check_cmd "$1"; then
    err "need '$1' (command not found)"
  fi
}

check_cmd() {
  command -v "$1" >/dev/null 2>&1
}

assert_nz() {
  if [ -z "$1" ]; then err "assert_nz $2"; fi
}

# Run a command that should never fail. If the command fails execution
# will immediately terminate with an error showing the failing
# command.
ensure() {
  if ! "$@"; then err "command failed: $*"; fi
}

header() {
  local _header
  _header=
  return "$_header"
}

download() {
  local _ciphersuites
  local _err
  local _status

  get_ciphersuites_for_curl
  _ciphersuites="$RETVAL"
  if [ -n "$_ciphersuites" ]; then
    _err=$(curl --proto '=https' --tlsv1.2 --ciphers "$_ciphersuites" --silent --show-error --fail --location "$1" --output "$2" 2>&1)
    _status=$?
  else
    echo "Warning: Not enforcing strong cipher suites for TLS, this is potentially less secure"
    if ! check_help_for "$3" curl --proto --tlsv1.2; then
      echo "Warning: Not enforcing TLS v1.2, this is potentially less secure"
      _err=$(curl --silent --show-error --fail --location "$1" --output "$2" 2>&1)
      _status=$?
    else
      _err=$(curl --proto '=https' --tlsv1.2 --silent --show-error --fail --location "$1" --output "$2" 2>&1)
      _status=$?
    fi
  fi
  if [ -n "$_err" ]; then
    echo "$_err" >&2
    if echo "$_err" | grep -q 404$; then
      err "installer for platform '$3' not found, this may be unsupported"
    fi
  fi
  return $_status
}

check_help_for() {
  local _arch
  local _cmd
  local _arg
  _arch="$1"
  shift
  _cmd="$1"
  shift

  local _category
  if "$_cmd" --help | grep -q 'For all options use the manual or "--help all".'; then
    _category="all"
  else
    _category=""
  fi

  case "$_arch" in

    *darwin*)
      if check_cmd sw_vers; then
        case $(sw_vers -productVersion) in
          10.*)
            # If we're running on macOS, older than 10.13, then we always
            # fail to find these options to force fallback
            if [ "$(sw_vers -productVersion | cut -d. -f2)" -lt 13 ]; then
              # Older than 10.13
              echo "Warning: Detected macOS platform older than 10.13"
              return 1
            fi
            ;;
          11.*)
            # We assume Big Sur will be OK for now
            ;;
          *)
            # Unknown product version, warn and continue
            echo "Warning: Detected unknown macOS major version: $(sw_vers -productVersion)"
            echo "Warning TLS capabilities detection may fail"
            ;;
        esac
      fi
      ;;

  esac

  for _arg in "$@"; do
    if ! "$_cmd" --help "$_category" | grep -q -- "$_arg"; then
      return 1
    fi
  done

  true # not strictly needed
}

# Return cipher suite string specified by user, otherwise return strong TLS 1.2-1.3 cipher suites
# if support by local tools is detected. Detection currently supports these curl backends:
# GnuTLS and OpenSSL (possibly also LibreSSL and BoringSSL). Return value can be empty.
get_ciphersuites_for_curl() {
  local _openssl_syntax="no"
  local _gnutls_syntax="no"
  local _backend_supported="yes"
  if curl -V | grep -q ' OpenSSL/'; then
    _openssl_syntax="yes"
  elif curl -V | grep -iq ' LibreSSL/'; then
    _openssl_syntax="yes"
  elif curl -V | grep -iq ' BoringSSL/'; then
    _openssl_syntax="yes"
  elif curl -V | grep -iq ' GnuTLS/'; then
    _gnutls_syntax="yes"
  else
    _backend_supported="no"
  fi

  local _args_supported="no"
  if [ "$_backend_supported" = "yes" ]; then
    # "unspecified" is for arch, allows for possibility old OS using macports, homebrew, etc.
    if check_help_for "notspecified" "curl" "--tlsv1.2" "--ciphers" "--proto"; then
      _args_supported="yes"
    fi
  fi

  local _cs=""
  if [ "$_args_supported" = "yes" ]; then
    if [ "$_openssl_syntax" = "yes" ]; then
      _cs=$(get_strong_ciphersuites_for "openssl")
    elif [ "$_gnutls_syntax" = "yes" ]; then
      _cs=$(get_strong_ciphersuites_for "gnutls")
    fi
  fi

  RETVAL="$_cs"
}

# Return strong TLS 1.2-1.3 cipher suites in OpenSSL or GnuTLS syntax. TLS 1.2
# excludes non-ECDHE and non-AEAD cipher suites. DHE is excluded due to bad
# DH params often found on servers (see RFC 7919). Sequence matches or is
# similar to Firefox 68 ESR with weak cipher suites disabled via about:config.
# $1 must be openssl or gnutls.
get_strong_ciphersuites_for() {
  if [ "$1" = "openssl" ]; then
    # OpenSSL is forgiving of unknown values, no problems with TLS 1.3 values on versions that don't support it yet.
    echo "TLS_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_256_GCM_SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384"
  elif [ "$1" = "gnutls" ]; then
    # GnuTLS isn't forgiving of unknown values, so this may require a GnuTLS version that supports TLS 1.3 even if wget doesn't.
    # Begin with SECURE128 (and higher) then remove/add to build cipher suites. Produces same 9 cipher suites as OpenSSL but in slightly different order.
    echo "SECURE128:-VERS-SSL3.0:-VERS-TLS1.0:-VERS-TLS1.1:-VERS-DTLS-ALL:-CIPHER-ALL:-MAC-ALL:-KX-ALL:+AEAD:+ECDHE-ECDSA:+ECDHE-RSA:+AES-128-GCM:+CHACHA20-POLY1305:+AES-256-GCM"
  fi
}

main "$@" || exit 1
