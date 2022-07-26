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