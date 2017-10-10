$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

if ((go version) -NotLike "*go1.9*") {
  echo "Must have go 1.9"
  exit 1
}

$ROOTDIR=(Split-Path -Parent $PSScriptRoot)
$env:GOPATH=$ROOTDIR

go run "$ROOTDIR/src/create/main.go" $ROOTDIR
