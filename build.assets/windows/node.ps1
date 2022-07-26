function Install-Node {
    <#
    .SYNOPSIS
        Downloads ands installs node
    #>
    [CmdletBinding()]
    param(
        [string] $ToolchainDir,
        [string] $NodeVersion
    )
    begin {
        $NodeZipfile = "node-$NodeVersion-win-x64.zip"
        Invoke-WebRequest -Uri https://nodejs.org/download/release/v$NodeVersion/node-v$NodeVersion-win-x64.zip `-OutFile $NodeZipfile
        Expand-Archive -Path $NodeZipfile -DestinationPath $ToolchainDir
        $Env:Path = "$Env:Path;$ToolchainDir/node-v$NodeVersion"
        corepack enable yarn
    }
}

function Enable-Node {
    <#
    .SYNOPSIS
        Downloads ands installs node
    #>
    [CmdletBinding()]
    param(
        [string] $ToolchainDir,
        [string] $NodeVersion
    )
    begin {
        $Env:Path = "$Env:Path;$ToolchainDir/node-v$NodeVersion"
    }
}
