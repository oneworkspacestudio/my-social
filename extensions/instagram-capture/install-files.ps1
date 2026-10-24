param(
 [Parameter(Mandatory=$true)][string]$Zip,
 [Parameter(Mandatory=$true)][ValidatePattern('^[a-fA-F0-9]{64}$')][string]$SHA256,
 [string]$Destination=(Join-Path ([Environment]::GetFolderPath('LocalApplicationData')) 'MySocial\InstagramCapture')
)
$ErrorActionPreference='Stop'
$source=(Resolve-Path -LiteralPath $Zip).Path
if((Get-FileHash -LiteralPath $source -Algorithm SHA256).Hash -ine $SHA256){throw 'Extension ZIP checksum does not match the published release'}
$target=[IO.Path]::GetFullPath($Destination).TrimEnd('\')
$parent=Split-Path $target -Parent
if(-not [IO.Path]::IsPathRooted($target) -or (Split-Path $target -Leaf) -ne 'InstagramCapture'){throw 'Choose a dedicated InstagramCapture installation folder'}
for($p=$target;$p;$p=Split-Path $p -Parent){if((Test-Path -LiteralPath $p) -and ((Get-Item -LiteralPath $p -Force).Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Linked installation paths are not supported'};if((Split-Path $p -Parent) -eq $p){break}}
if(Test-Path -LiteralPath $target){throw 'An installation already exists. Keep it and use a separately verified update; nothing was overwritten.'}
New-Item -ItemType Directory -Path $parent -Force | Out-Null
Add-Type -AssemblyName System.IO.Compression.FileSystem
$archive=[IO.Compression.ZipFile]::OpenRead($source)
try{
 $total=0
 foreach($entry in $archive.Entries){$name=$entry.FullName;if($name -match '(^[/\\]|(^|[/\\])\.\.([/\\]|$)|:)' -or $name.Contains('\')){throw 'Unsafe extension ZIP member'};$total+=$entry.Length}
 if($total -gt 16MB -or $archive.Entries.Count -gt 200){throw 'Unexpected extension archive size'}
 if(-not($archive.Entries | Where-Object FullName -eq 'manifest.json')){throw 'The ZIP root must contain manifest.json'}
}finally{$archive.Dispose()}
[IO.Compression.ZipFile]::ExtractToDirectory($source,$target)
$manifest=[IO.File]::ReadAllText((Join-Path $target 'manifest.json'),[Text.Encoding]::UTF8) | ConvertFrom-Json
if($manifest.name -ne 'My Social Instagram Capture' -or $manifest.manifest_version -ne 3){throw 'Unexpected extension identity; review the extracted files'}
[IO.File]::WriteAllText((Join-Path $target '.mysocial-install.json'),(@{version=$manifest.version;zipSHA256=$SHA256.ToLowerInvariant();format='my-social-instagram-capture-v1'}|ConvertTo-Json))
Write-Output "Verified extension files installed at $target"
Write-Output 'In Chrome open chrome://extensions, enable Developer mode, click Load unpacked, and select that folder.'
