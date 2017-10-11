﻿$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

$rootdir=(Split-Path -Parent $PSScriptRoot)
$outfile = "$rootdir/bin/create.exe"

if ($env:DEV_ENV -ne $null -and $env:DEV_ENV -ne "") {
  if ((go version) -NotLike "*go1.9*") {
    echo "Must have go 1.9"
    exit 1
  }

  $env:GOPATH=$rootdir
  go build -o "$outfile" "$rootdir/src/create/main.go"
} else {
  $version=(cat "$rootdir/VERSION")
  $url="https://s3.amazonaws.com/windows2016fs/create-binaries/create-$version-windows-amd64.exe"

  mkdir -p "$rootdir/bin"
  $wc = New-Object net.webclient
  $wc.Downloadfile($url, $outfile)
}

& "$outfile" "$rootdir"
exit $LASTEXITCODE
