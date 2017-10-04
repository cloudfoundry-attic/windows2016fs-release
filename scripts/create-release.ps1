$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

if ((go version) -NotLike "*go1.9*") {
  echo "Must have go 1.9"
  exit 1
}
if ((bosh -v) -NotLike "*version 2*") {
  echo "Must have BOSH cli v2"
  exit 1
}

$ROOTDIR=(Split-Path -Parent $PSScriptRoot)

$env:GOPATH=$ROOTDIR

Push-Location $ROOTDIR
  $image_tag=(cat IMAGE_TAG)
  $image_name="cloudfoundry/windows2016fs"
  $output_dir="blobs/windows2016fs"
  mkdir -Force $output_dir

  go run src/oci-image/cmd/hydrate/main.go -image $image_name -outputDir $output_dir -tag $image_tag

  $release_version=(cat VERSION)
  bosh cr --version=$release_version
Pop-Location
