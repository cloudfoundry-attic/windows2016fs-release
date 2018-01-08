$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

# workaround for https://github.com/Microsoft/hcsshim/issues/155
Get-ComputeProcess | Stop-ComputeProcess -f

$layersDir = "c:\var\vcap\data\windows2016fs\layers"
$topLayer = (c:\var\vcap\packages\extract\extract.exe c:\var\vcap\packages\windows2016fs\windows2016fs.tgz $layersDir)
if ($LASTEXITCODE -ne 0) {
  Exit $LASTEXITCODE
}

$rootfsPackage="C:\var\vcap\packages\windows2016fs"
$rootfsSymlink="$rootfsPackage\rootfs"

if ((Test-Path $rootfsSymlink) -eq $true) {
  cmd.exe /c "rmdir $rootfsSymlink"
  if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to remove old symlink"
  }
}

cmd.exe /c "mklink /D $rootfsSymlink $topLayer"
if ($LASTEXITCODE -ne 0) {
  Write-Error "Failed to create symlink"
}
