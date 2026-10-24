param([string]$Go='go',[string]$Out='dist/release',[string]$Version='0.1.0-preview.6',[string]$Rsrc='')
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
Push-Location $root
try {
 $env:GOWORK='off'
 & $Go mod verify
 if($LASTEXITCODE){throw 'Module verification failed'}
 node web/scripts/icons.mjs
 if($LASTEXITCODE){throw 'Icon generation failed'}
 npm --prefix web run build
 if($LASTEXITCODE){throw 'Frontend build failed'}
 npm --prefix web test
 if($LASTEXITCODE){throw 'Frontend checks failed'}
 npm --prefix web run extension:build
 if($LASTEXITCODE){throw 'Extension build failed'}
 npm --prefix web run extension:test
 if($LASTEXITCODE){throw 'Extension tests failed'}
 & $Go test ./...
 if($LASTEXITCODE){throw 'Go tests failed'}
 & $Go vet ./...
 if($LASTEXITCODE){throw 'Go vet failed'}
 New-Item -ItemType Directory -Force $Out | Out-Null
 $Out=(Resolve-Path -LiteralPath $Out).Path
 if($Rsrc){& $Rsrc -ico docs/branding/my_social.ico -o cmd/my-social/rsrc_windows_amd64.syso}
 else {& $Go run github.com/akavel/rsrc@v0.10.2 -ico docs/branding/my_social.ico -o cmd/my-social/rsrc_windows_amd64.syso}
 if($LASTEXITCODE){throw 'Windows icon generation failed'}
 & $Go build -buildvcs=false -trimpath -ldflags "-s -w -X main.version=$Version" -o (Join-Path $Out 'MySocial.exe') ./cmd/my-social
 if($LASTEXITCODE){throw 'Windows executable build failed'}
 python scripts/notices.py --go $Go --out $Out
 if($LASTEXITCODE){throw 'Dependency notices failed'}
 Copy-Item -LiteralPath (Join-Path $Out 'THIRD-PARTY-LICENSES.txt') -Destination 'extensions/instagram-capture/dist/THIRD-PARTY-LICENSES.txt' -Force
 $extensionVersion=(Get-Content -LiteralPath 'extensions/instagram-capture/manifest.json' -Raw | ConvertFrom-Json).version
 $extensionZip=Join-Path (Split-Path $Out -Parent) ('my-social-instagram-capture-'+$extensionVersion+'.zip')
 Compress-Archive -Path 'extensions/instagram-capture/dist/*' -DestinationPath $extensionZip -Force
 Copy-Item -LiteralPath 'extensions/instagram-capture/install-files.ps1' -Destination (Join-Path (Split-Path $Out -Parent) 'install-instagram-capture.ps1') -Force
 foreach($file in @('LICENSE','NOTICE','README.md','SECURITY.md','CHANGELOG.md','compatibility.json')){Copy-Item -LiteralPath $file -Destination $Out -Force}
 Copy-Item -LiteralPath docs -Destination $Out -Recurse -Force
 $commit=(git rev-parse HEAD).Trim()
 [IO.File]::WriteAllText((Join-Path $Out 'build-info.json'),(@{version=$Version;commit=$commit;runtime='windows-amd64';apphost='31171aa1970e86f1a4130af45dd32bc3b776eeb4';status='evaluation-preview'}|ConvertTo-Json))
 $zip=Join-Path (Split-Path $Out -Parent) ('my-social-'+$Version+'-windows-amd64.zip')
 Compress-Archive -Path (Join-Path $Out '*') -DestinationPath $zip -Force
 $source=Join-Path (Split-Path $Out -Parent) 'my-social-source.zip'
 git archive --format=zip --output=$source HEAD
 if($LASTEXITCODE){throw 'Source archive failed'}
 $hashes=foreach($file in @((Join-Path $Out 'MySocial.exe'),$zip,$source,$extensionZip)){((Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant())+'  '+(Split-Path $file -Leaf)}
 [IO.File]::WriteAllLines((Join-Path (Split-Path $Out -Parent) 'SHA256SUMS.txt'),$hashes)
 Write-Output ('Built '+$zip)
}finally{Pop-Location}
