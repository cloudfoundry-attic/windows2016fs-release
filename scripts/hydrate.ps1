$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

if ((go version) -NotLike "*go1.9*") {
  echo "Must have go 1.9"
  exit 1
}

$ROOTDIR=(Split-Path -Parent $PSScriptRoot)

$env:GOPATH=$ROOTDIR

Push-Location $ROOTDIR
  $image_tag=(cat IMAGE_TAG)
  $image_name="cloudfoundry/windows2016fs"
  $output_dir="blobs/windows2016fs"
  mkdir -Force $output_dir

  go run src/hydrate/cmd/hydrate/main.go -image $image_name -outputDir $output_dir -tag $image_tag
Pop-Location
