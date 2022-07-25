// Copyright 2021 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import "path"

const (
	windowsToolchainDir = `$TEMP\$DRONE_BUILD_NUMBER-$DRONE_BUILD_CREATED\toolchains`
)

func newWindowsPipeline(name string) pipeline {
	p := newExecPipeline(name)
	p.Workspace.Path = path.Join("C:/Drone/Workspace", name)
	p.Concurrency.Limit = 1
	p.Platform = platform{OS: "windows", Arch: "amd64"}
	return p
}

func windowsPushPipeline() pipeline {
	p := newWindowsPipeline("push-build-native-windows-amd64")
	p.Trigger = trigger{
		Event:  triggerRef{Include: []string{"push"}, Exclude: []string{"pull_request"}},
		Branch: triggerRef{Include: []string{"master", "branch/*", "tcsc/build-windows*"}},
		Repo:   triggerRef{Include: []string{"gravitational/*"}},
	}

	perBuildWorkspace := "$Env:WORKSPACE_DIR/$Env:DRONE_BUILD_NUMBER"
	perBuildTeleportSrc := perBuildWorkspace + "/go/src/github.com/gravitational/teleport"
	perBuildWebappsSrc := perBuildWorkspace + "/go/src/github.com/gravitational/webapps"

	p.Steps = []step{
		installWindowsNodeToolchainStep(p.Workspace.Path),
		{
			Name: "Check out code",
			Environment: map[string]value{
				"WORKSPACE_DIR":      {raw: p.Workspace.Path},
				"GITHUB_PRIVATE_KEY": {fromSecret: "GITHUB_PRIVATE_KEY"},
			},
			Commands: []string{
				`$ErrorActionPreference = 'Stop'`,
				`$TeleportSrc = ` + perBuildTeleportSrc,
				`New-Item -Path $TeleportSrc -ItemType Directory | Out-Null`,
				`cd $TeleportSrc`,
				`git clone https://github.com/gravitational/${DRONE_REPO_NAME}.git .`,
				`git checkout ${DRONE_TAG:-$DRONE_COMMIT}`,
				`$WebappsSrc = ` + perBuildWebappsSrc,
				`cd $WebappsSrc`,
				`git clone https://github.com/gravitational/webapps.git .`,
				`git checkout $(go run $TeleportSrc/build.assets/tooling/cmd/get-webapps-version)`,
			},
		},
		cleanUpWindowsToolchainsStep(p.Workspace.Path),
	}

	return p
}

func installWindowsNodeToolchainStep(workspacePath string) step {
	return step{
		Name:        "Install Node Toolchain",
		Environment: map[string]value{"WORKSPACE_DIR": {raw: workspacePath}},
		Commands: []string{
			`$global:ProgressPreference = 'SilentlyContinue'`,
			`$ErrorActionPreference = 'Stop'`,
			`$NodeVersion = $(make -C $WORKSPACE_DIR/go/src/github.com/gravitational/teleport/build.assets print-node-version)`,
			`$NodeZipfile = "node-$NodeVersion-win-x64.zip"`,
			`Invoke-WebRequest -Uri https://nodejs.org/download/release/v$NodeVersion/node-v$NodeVersion-win-x64.zip -OutFile $NodeZipfile`,
			`Expand-Archive -Path $NodeZip -DestinationPath ` + windowsToolchainDir,
			`$Env:Path = $Env:Path;` + windowsToolchainDir + `node-v$NodeVersion`,
			`corepack enable yarn`,
		},
	}
}

func cleanUpWindowsToolchainsStep(workspacePath string) step {
	return step{
		Name:        "Clean up toolchains (post)",
		Environment: map[string]value{"WORKSPACE_DIR": {raw: workspacePath}},
		When: &condition{
			Status: []string{"success", "failure"},
		},
		Commands: []string{
			`Remove-Item -Recurse -Path ` + windowsToolchainDir,
		},
	}
}
