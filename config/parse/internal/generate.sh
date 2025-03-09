#!/usr/bin/env bash
#
# This script generates a lexer for the Config parser.
#  • The only time this script needs to be run is when the lexer rules change:
#    · config/parse/parser.go -> func LexerDefinition()
#
#  • Do not run directly; this script depends on env vars set by `go generate`.

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

  # gum will only work with executables (not functions),
  # so we just call ourself with the name of the function to run.
  gum spin \
    --title="generating lexer: ${lexer_src_path}" \
    --spinner="minidot" \
    --show-output \
    -- "${0}" "${@}"
}

set -oo errexit pipefail
case ${#} in
  (0) _init _run ;;
  (*) "${1}" "${@:2}" ;;
esac
