# windows2016fs-release

## Using this release

Due to limitations in distributing the Microsoft container images, this release does not actually have any final releases. As such, building and versioning this release is slightly unconventional. 

`scripts/create-release` and `scripts/create-release.ps1` can be used to create a release which can be uploaded to a bosh director. This release will have a correct version and will use the correct `cloudfoundry/windows2016fs` container image.

## Usage

### Windows
```
./scripts/create-release.ps1 -tarball {{file.tgz}}
```

### Linux
```
./scripts/create-release --tarball {{file.tgz}}
```

If you are running in dev mode, set the `DEV_ENV` environment variable to `true`.

