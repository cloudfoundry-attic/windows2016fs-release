Param (
  [string]$tarball=""
)

$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

$rootdir=(Split-Path -Parent $PSScriptRoot)
$outfile = "$rootdir/bin/create.exe"

$env:GOPATH=$rootdir
go build -o "$outfile" "$rootdir/src/create/main.go"
if ($LASTEXITCODE -ne 0) {
  Exit $LASTEXITCODE
}

if ($tarball -ne "") {
  echo "Will write tarball to $tarball"
  & "$outfile" --releaseDir "$rootdir" --tarball "$tarball"
} else {
  echo "NO TARBALL"
  & "$outfile" --releaseDir "$rootdir"
}

exit $LASTEXITCODE
