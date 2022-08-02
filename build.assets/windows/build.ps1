function Enable-Git {
    <#
    .SYNOPSIS
        Configures git for accessing (possibly private) repos
    #>
    [CmdletBinding()]
    param(
        [string] $Workspace,
        [string] $PrivateKey
    )
    begin {
        $SSHDir = "$Workspace/.ssh"
        New-Item -Path "$SSHDir" -ItemType Directory | Out-Null
        $PrivateKey | Out-File -Encoding ascii "$SSHDir/id_rsa"
        Invoke-WebRequest "https://api.github.com/meta" -UseBasicParsing `
            | ConvertFrom-JSON `
            | Select-Object -ExpandProperty "ssh_keys" `
            | ForEach-Object {"github.com $_"} `
            | Out-File -Encoding ASCII "$SSHDir/known_hosts"
        $SSHCmd = "ssh -i $SSHDir/id_rsa -o UserKnownHostsFile=$SSHDir/known_hosts -F/dev/null"
        $Env:GIT_SSH_COMMAND = $SSHCmd
    }
}

function Reset-Git {
[CmdletBinding()]
param(
    [string] $Workspace
)
begin {
    Remove-Item -Recurse -Path "$Workspace/.ssh"
}
}

function Install-Go {
    <#
    .SYNOPSIS
        Downloads ands installs go
    #>
    [CmdletBinding()]
    param(
        [string] $ToolchainDir,
        [string] $GoVersion
    )
    begin {
        $GoDownloadUrl = "https://go.dev/dl/go$GoVersion.windows-amd64.zip"
        $GoInstallZip = "go$GoVersion.windows-amd64.zip"
        Invoke-WebRequest -Uri $GoDownloadUrl -OutFile $GoInstallZip
        Expand-Archive -Path $GoInstallZip -DestinationPath $ToolchainDir
        $Env:Path = "$Env:Path;$ToolchainDir/go/bin"
    }
}

function Enable-Go {
    <#
    .SYNOPSIS
        Adds go to the environment 
    #>
    [CmdletBinding()]
    param(
        [string] $ToolchainDir,
        [string] $GoVersion
    )
    begin {
        $Env:Path = "$Env:Path;$ToolchainDir/go/bin"
    }
}

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
        Invoke-WebRequest -Uri https://nodejs.org/download/release/v$NodeVersion/node-v$NodeVersion-win-x64.zip -OutFile $NodeZipfile
        Expand-Archive -Path $NodeZipfile -DestinationPath $ToolchainDir
        $Env:Path = "$Env:Path;$ToolchainDir/node-v$NodeVersion-win-x64"
        npm config set msvs_version 2017
        corepack enable yarn
    }
}

function Enable-Node {
    <#
    .SYNOPSIS
        Adds node to the environment 
    #>
    [CmdletBinding()]
    param(
        [string] $ToolchainDir,
        [string] $NodeVersion
    )
    begin {
        $Env:Path = "$Env:Path;$ToolchainDir/node-v$NodeVersion-win-x64"
    }
}
