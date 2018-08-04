#!/usr/bin/env bash

set -e

out="${1:-/dev/stdout}"
tmpfile="$(mktemp -t "$(basename "$0").XXXXXX")"

if ! which interfacer ; then
  # dependencies
  go1.11beta2 get github.com/rjeczalik/interfaces
  go1.11beta2 get github.com/rjeczalik/interfaces/cmd/interfacer
fi

if [ -e ./shell ]; then
  cd ./shell
fi


(
  interfacer -for '".".Shell' -as shell.Interface |
    sed 's/shell\.//g' |
    sed 's/for ".".Shell/for Shell/g' |
    sed 's/"."//g' |
    sed 's/\[\]interface{}/...interface{}/g' |
    gofmt

  cat <<EOS

// ensure compatible w/ Shell
var mockShellInterface Interface = &MockShell{}
var shellInterface Interface = &Shell{}
EOS
) > "$tmpfile"

cp "$tmpfile" "$out"
