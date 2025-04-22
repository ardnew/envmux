#!/usr/bin/env bash
#
# This script generates documentation on the grammar of the configuration file
# in two different formats:
#
#   1. A definition in eBNF form (ASCII text)
#   2. A Web page (html, js, css) with railroad diagrams in SVG format

if ! rr=$( type -P railroad ); then
  inst=( 'go' 'install' 'github.com/alecthomas/participle/cmd/railroad@latest' )
  warn="Railroad diagram generator not found: ${rr}"
  cont="Install (${inst[*]})"
  quit='Quit'

  if gum=$( type -P gum ); then
    curs='•> '
    args=(
      ordered
      header="${warn}"
      cursor="${curs}"
      label-delimiter=':'
      limit=1
      no-show-help
      no-strip-ansi
      selected="${quit}"
    )
    ans=$( "${gum}" choose "${args[@]/#/--}" "${cont}:Y" "${quit}:N" )
  else
    echo "${warn}" >&2
    read -r -n 1 -p "${cont}? [y/N] " ans
  fi

  [[ ${ans,,} == "y" ]] || exit 1
  "${inst[@]}"
fi

self=$( realpath -qe "${0}" )
path=${self%/*}

go run "${path}/grammar" |& tee "${path}/grammar/internal/ebnf.asc"
pushd "${path}/grammar/internal" &>/dev/null || exit 1
railroad -w -o "grammar.html" < "${path}/grammar/internal/ebnf.asc"
popd &>/dev/null || exit 1

echo "Generated grammar documentation:"
echo "  • EBNF: ${path}/grammar/internal/ebnf.asc"
echo "  • HTML: ${path}/grammar/internal/grammar.html"
