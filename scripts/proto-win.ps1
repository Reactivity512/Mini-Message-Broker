$ErrorActionPreference = "Stop"
$ProtoDir = "api/proto"
$ProtoOut = "internal/delivery/grpc/pb"
$ProtocVersion = "33.5"
$ProtocZip = "protoc-$ProtocVersion-win64.zip"
$ProtocUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v$ProtocVersion/$ProtocZip"
$ToolsDir = ".tools"
$ProtocDir = "$ToolsDir/protoc"

if (Get-Command protoc -ErrorAction SilentlyContinue) {
    Write-Host "Using protoc from PATH"
    $protoc = "protoc"
} else {
    if (-not (Test-Path "$ProtocDir/bin/protoc.exe")) {
        Write-Host "Downloading protoc v$ProtocVersion..."
        New-Item -ItemType Directory -Force -Path $ProtocDir | Out-Null
        Invoke-WebRequest -Uri $ProtocUrl -OutFile "$ProtocDir/$ProtocZip" -UseBasicParsing
        Expand-Archive -Path "$ProtocDir/$ProtocZip" -DestinationPath $ProtocDir -Force
        Remove-Item "$ProtocDir/$ProtocZip" -Force -ErrorAction SilentlyContinue
        # zip has bin/protoc.exe and include/ at root of archive
        if (-not (Test-Path "$ProtocDir/bin/protoc.exe")) {
            $sub = Get-ChildItem $ProtocDir -Directory | Select-Object -First 1
            if ($sub) { Move-Item "$($sub.FullName)/*" $ProtocDir -Force; Remove-Item $sub.FullName -Force }
        }
    }
    $protoc = "$ProtocDir/bin/protoc.exe"
    if (-not (Test-Path $protoc)) { $protoc = "$ProtocDir/protoc.exe" }
    $protoc = (Resolve-Path $protoc).Path
    Write-Host "Using protoc: $protoc"
}

New-Item -ItemType Directory -Force -Path $ProtoOut | Out-Null
& $protoc --go_out=$ProtoOut --go_opt=paths=source_relative `
    --go-grpc_out=$ProtoOut --go-grpc_opt=paths=source_relative `
    -I $ProtoDir "$ProtoDir/*.proto"
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Generated files in $ProtoOut"
