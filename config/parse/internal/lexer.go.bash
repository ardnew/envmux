#!/usr/bin/env bash
#
# This script generates a lexer for the Config parser.
#  • Do NOT run this script directly!
#    · It depends on env vars set by `go generate` (shown below).
#  • Run this script:
#    · When -> The lexer rules¹ have changed.
#    · How² -> `go generate ./config/parse/...`
#
# [¹]: var LexerDefinition is defined in "config/parse/parser.go".
# [²]: Run this command from the module root directory (containing "go.mod").

gen-rules() {
  "${go_bin_path}" run "${marshal_pkg_path}" "${lexer_rules_path}"
}

gen-lexer() {
  "${participle_bin_path}" gen lexer --name "${lexer_ident_prefix}" "${GOPACKAGE}" < "${lexer_rules_path}" |
    perl -ple 'print "$ENV{generate_directive}" if $. == 1' | gofmt -s | tee "${lexer_src_path}" &>/dev/null
}

gen-clean() {
  rm -f "${lexer_rules_path}"
}

_run() {
  gen-rules
  gen-lexer
  gen-clean
}

_init() {
  declare -rx lexer_ident_prefix="Config"

  declare -rx generate_directive=$( sed -n "${GOLINE}p" "${GOFILE}" )

  declare -rx lexer_src_path=$( realpath -qe "${GOFILE}" )
  declare -rx internal_pkg_path="${lexer_src_path%/*}/internal"
  declare -rx marshal_pkg_path="./$( realpath -qe "${internal_pkg_path}/marshal" --relative-to="${PWD}" )"
  declare -rx lexer_rules_path="${internal_pkg_path}/rules.json"

  declare -rx participle_bin_path=$( type -P participle )
  declare -rx go_bin_path=$( type -P go )

  # if installed, use `gum` to render a spinner while generating.
  # otherwise, echo and exec like a luddite.

  # gum cannot resolve functions, as it must invoke an executable.
  #
  # so we use ourself as the executable (${0}), passing the function and its
  # arguments as arguments to ourself.
  #
  # this works only because the first invocation of this script must be made
  # with no arguments. in that case, we make the recursive call. otherwise, we
  # invoke the function passed as the first argument.
  #
  # thus, do not ever pass "_init" as the first argument to this script.
  # your PC will explode.
  declare -r status="generating lexer: ${lexer_src_path}"
  ! gum=$( type -P gum ) ||
    exec "${gum}" spin \
      --title="${status}" \
      --spinner="minidot" \
      --show-output \
      -- "${0}" "${@}"

  echo "${status}"
  exec "${0}" "${@}"
}

set -oo errexit pipefail

# these will likely be undefined if running this script directly.
{
  declare -r undef='error: variable must be set'
  : ${GOPACKAGE?${undef}}
  : ${GOLINE?${undef}}
  : ${GOFILE?${undef}}
}

# ===[ main ]===

case ${#} in
  (0) _init _run ;;      # no argument(s) -> invoked by `go generate`.
  (*) "${1}" "${@:2}" ;; # argument(s) given -> recursive call from _init()
esac
