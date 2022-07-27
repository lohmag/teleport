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
	perBuildWorkspace   = `$Env:WORKSPACE_DIR/$Env:DRONE_BUILD_NUMBER`
	windowsToolchainDir = perBuildWorkspace + `/toolchains`
	perBuildTeleportSrc = perBuildWorkspace + "/go/src/github.com/gravitational/teleport"
	perBuildWebappsSrc  = perBuildWorkspace + "/go/src/github.com/gravitational/webapps"
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

	p.Steps = []step{
		{
			Name: "Check out Teleport",
			Environment: map[string]value{
				"WORKSPACE_DIR": {raw: p.Workspace.Path},
			},
			Commands: []string{
				`$ErrorActionPreference = 'Stop'`,
				`$Env:GOCACHE = "` + perBuildWorkspace + `/gocache"`,
				`$TeleportSrc = "` + perBuildTeleportSrc + `"`,
				`$TeleportRev = "$Env:DRONE_COMMIT"`, // need to allow override for tag
				`New-Item -Path $TeleportSrc -ItemType Directory | Out-Null`,
				`cd $TeleportSrc`,
				`git clone https://github.com/gravitational/${DRONE_REPO_NAME}.git .`,
				`git checkout $TeleportRev`,
				`$WebappsSrc = "` + perBuildWebappsSrc + `"`,
				`New-Item -Path $WebappsSrc -ItemType Directory | Out-Null`,
				`cd $WebappsSrc`,
				`git clone https://github.com/gravitational/webapps.git .`,
				`git checkout $(go run $TeleportSrc/build.assets/tooling/cmd/get-webapps-version/main.go)`,
			},
		}, {
			Name: "Checkout Submodules",
			Environment: map[string]value{
				"WORKSPACE_DIR":      {raw: p.Workspace.Path},
				"GITHUB_PRIVATE_KEY": {fromSecret: "GITHUB_PRIVATE_KEY"},
			},
			Commands: []string{
				`$Workspace = "` + perBuildWorkspace + `"`,
				`$TeleportSrc = "` + perBuildTeleportSrc + `"`,
				`. "$TeleportSrc/build.assets/windows/build.ps1"`,
				`Enable-Git -Workspace $Workspace -PrivateKey $Env:GITHUB_PRIVATE_KEY`,
				`cd $TeleportSrc`,
				`git submodule update --init e`,
				`git submodule update --init --recursive webassets`,
				`Reset-Git -Workspace $Workspace`,
			},
		},
		installWindowsNodeToolchainStep(p.Workspace.Path),
		installWindowsGoToolchainStep(p.Workspace.Path),

		cleanUpWindowsWorkspaceStep(p.Workspace.Path),
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
			`$TeleportSrc = "` + perBuildTeleportSrc + `"`,
			`. "$TeleportSrc/build.assets/windows/build.ps1"`,
			// We can't use make, as there are too many posix dependencies to
			// abstract away right now, so instead of `$(make -C $TeleportSrc/build.assets print-node-version)`,
			// we will just hardcode it for now
			`$NodeVersion = "16.13.2"`,
			`Install-Node -NodeVersion $NodeVersion -ToolchainDir "` + windowsToolchainDir + `"`,
		},
	}
}

func installWindowsGoToolchainStep(workspacePath string) step {
	return step{
		Name:        "Install Go Toolchain",
		Environment: map[string]value{"WORKSPACE_DIR": {raw: workspacePath}},
		Commands: []string{
			`$global:ProgressPreference = 'SilentlyContinue'`,
			`$ErrorActionPreference = 'Stop'`,
			`$TeleportSrc = "` + perBuildTeleportSrc + `"`,
			`. "$TeleportSrc/build.assets/windows/build.ps1"`,
			// We can't use make, as there are too many posix dependencies to
			// abstract away right now, so instead of `$(make -C $TeleportSrc/build.assets print-go-version)`,
			// we will just hardcode it for now
			`$GoVersion = "1.18.3"`,
			`Install-Go -GoVersion $GoVersion -ToolchainDir "` + windowsToolchainDir + `"`,
		},
	}
}

func cleanUpWindowsWorkspaceStep(workspacePath string) step {
	return step{
		Name:        "Clean up workspace (post)",
		Environment: map[string]value{"WORKSPACE_DIR": {raw: workspacePath}},
		When: &condition{
			Status: []string{"success", "failure"},
		},
		Commands: []string{
			`Remove-Item -Recurse -Force -Path "$Env:WORKSPACE_DIR/$Env:DRONE_BUILD_NUMBER"`,
		},
	}
}
