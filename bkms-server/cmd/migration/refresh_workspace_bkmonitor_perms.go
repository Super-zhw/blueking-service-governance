/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 服务治理 (BlueKing Service Governance) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *  http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

package migration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/bkintegrations/bkiam"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/bkintegrations/bkiam/role"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/bkintegrations/bkiam/scope"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/config"
	log "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/logging"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/workspace"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/perm"
	storereg "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/server/registry"
)

// NewRefreshWorkspaceBkmonitorPermsCmd 批量刷新所有 Ready 状态 workspace 的 bkmonitor
// 权限范围（含 MCP actions）到 IAM grade manager 及用户组。
// 支持 --dry-run 预览待变更内容。
func NewRefreshWorkspaceBkmonitorPermsCmd() *cobra.Command {
	var srvCfg string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "refresh_workspace_bkmonitor_perms",
		Short: "Refresh bkmonitor permission scopes (including MCP actions) for all workspaces to IAM user groups",
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := cmd.Context()
			cfg, err := config.Load(ctx, srvCfg)
			if err != nil {
				log.Fatalf("load config: %v", err)
			}
			if err = log.InitDefaultLogger(cfg.Logging); err != nil {
				log.Fatalf("init logger: %v", err)
			}

			database.InitClient(ctx, cfg.Mongo)
			storereg.Init(ctx)

			wsStore := storereg.G().WorkspaceStore

			if dryRun {
				dryRunRefreshWorkspaceBkmonitorPerms(ctx, wsStore)
			} else {
				permMgr := perm.NewManager()
				refreshWorkspaceBkmonitorPerms(ctx, wsStore, permMgr)
			}
		},
	}

	cmd.Flags().StringVar(&srvCfg, "srvCfg", "", "server config file")
	cmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "list pending changes without actually executing IAM permission refresh")
	_ = cmd.MarkFlagRequired("srvCfg")

	return cmd
}

// dryRunRefreshWorkspaceBkmonitorPerms 以 dry-run 模式列出所有待变更的 workspace 及其详细信息。
func dryRunRefreshWorkspaceBkmonitorPerms(
	ctx context.Context,
	wsStore workspace.WorkspaceStore,
) {
	readyState := workspace.StateReady
	workspaces, err := wsStore.List(ctx, &workspace.ListOptions{State: &readyState})
	if err != nil {
		log.Fatalf("list workspaces: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("[DRY-RUN] BKMonitor Permission Refresh Preview")
	fmt.Printf("========================================\n\n")
	fmt.Printf("Total workspaces to process: %d\n\n", len(workspaces))

	var willProcess, willSkip int

	for i, ws := range workspaces {
		fmt.Printf("--- [%d/%d] Workspace: %s (%s) ---\n", i+1, len(workspaces), ws.ID, ws.DisplayName)

		if ws.BkSystems.BkMonitorProjectID == "" {
			fmt.Printf("  Status: SKIP (BkMonitorProjectID is empty)\n\n")
			willSkip++
			continue
		}

		willProcess++
		data := buildWorkspaceData(ws)

		fmt.Printf("  Status: WILL_REFRESH\n")
		fmt.Printf("  BKMonitor: SpaceID=%s\n", data.BKMonitor.SpaceID)
		if data.BKLog != nil {
			fmt.Printf("  BKLog:     SpaceID=%s\n", data.BKLog.SpaceID)
		}
		if data.BKCI != nil {
			fmt.Printf("  BKCI:     ProjectID=%s\n", data.BKCI.ProjectID)
		}
		if data.BCS != nil {
			fmt.Printf("  BCS:      ProjectID=%s\n", data.BCS.ProjectID)
		}
		if data.BKRepo != nil {
			fmt.Printf("  BKRepo:   ProjectID=%s\n", data.BKRepo.ProjectID)
		}

		fmt.Println("  BKMonitor actions to grant (by role):")
		for _, roleCode := range append(
			[]string{role.BuiltinRoleCode.Admin},
			role.WorkspaceScopeBuiltinRoles...,
		) {
			g := scope.BKMonitorRoleScopesGenerator{
				SpaceID:     data.BKMonitor.SpaceID,
				SpaceName:   data.BKMonitor.SpaceName,
				TplRoleCode: roleCode,
			}
			scopes := g.Generate()
			fmt.Printf("    [%s] %d scope blocks, actions: ", roleCode, len(scopes))
			for si, s := range scopes {
				if si > 0 {
					fmt.Print(" | ")
				}
				for ai, a := range s.Actions {
					if ai > 0 {
						fmt.Print(", ")
					}
					fmt.Print(a.ID)
				}
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Println("========================================")
	fmt.Printf("[DRY-RUN] Summary: total=%d, will_process=%d, will_skip=%d\n", len(workspaces), willProcess, willSkip)
	fmt.Println("========================================")

	if willProcess > 0 {
		for _, ws := range workspaces {
			if ws.BkSystems.BkMonitorProjectID == "" {
				continue
			}
			data := buildWorkspaceData(ws)
			jsonBytes, _ := json.MarshalIndent(data, "", "  ")
			fmt.Printf("\n[DRY-RUN] Sample WorkspaceData (workspace=%s):\n%s\n", ws.ID, string(jsonBytes))
			break
		}
	}
}

// refreshWorkspaceBkmonitorPerms 遍历所有 Ready 状态的 workspace，逐个刷新 bkmonitor 权限。
func refreshWorkspaceBkmonitorPerms(
	ctx context.Context,
	wsStore workspace.WorkspaceStore,
	permMgr perm.Manager,
) {
	readyState := workspace.StateReady
	workspaces, err := wsStore.List(ctx, &workspace.ListOptions{State: &readyState})
	if err != nil {
		log.Fatalf("list workspaces: %v", err)
	}

	var total, success, skipped, failed int
	total = len(workspaces)

	for _, ws := range workspaces {
		if ws.BkSystems.BkMonitorProjectID == "" {
			log.Warnf(ctx, "workspace %s has no BkMonitorProjectID, skipping", ws.ID)
			skipped++
			continue
		}

		data := buildWorkspaceData(ws)

		if err = permMgr.UpdateWorkspaceAdmin(ctx, data); err != nil {
			log.Errorf(ctx, "update workspace admin for %s failed: %v", ws.ID, err)
			failed++
			continue
		}

		if err = permMgr.UpdateWorkspaceScopeBuiltinRoles(ctx, data); err != nil {
			log.Errorf(ctx, "update workspace builtin roles for %s failed: %v", ws.ID, err)
			failed++
			continue
		}

		log.Infof(ctx, "workspace %s bkmonitor perms refreshed successfully", ws.ID)
		success++
	}

	log.Infof(
		ctx,
		"refresh_workspace_bkmonitor_perms completed: total=%d, success=%d, skipped=%d, failed=%d",
		total, success, skipped, failed,
	)
}

// buildWorkspaceData 根据 workspace 信息构造 WorkspaceData，填充所有平台字段。
func buildWorkspaceData(ws workspace.Workspace) bkiam.WorkspaceData {
	data := bkiam.WorkspaceData{
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.DisplayName,
		BKMonitor: &bkiam.BKMonitorOptions{
			SpaceID:   ws.BkSystems.BkMonitorProjectID,
			SpaceName: ws.DisplayName,
		},
	}

	// 填充 BKLog（与 BKMonitor 共用同一个 project ID）
	if ws.BkSystems.BkLogProjectID != "" {
		data.BKLog = &bkiam.BKLogOptions{
			SpaceID:   ws.BkSystems.BkLogProjectID,
			SpaceName: ws.DisplayName,
		}
	}

	// 填充 BKCI
	if ws.BkSystems.BkCIProjectID != "" {
		data.BKCI = &bkiam.BKCIOptions{
			ProjectID:   ws.BkSystems.BkCIProjectID,
			ProjectName: ws.BkSystems.BkCIProjectID,
		}
	}

	// 填充 BCS
	if ws.BkSystems.BkBCSProjectID != "" {
		data.BCS = &bkiam.BCSOptions{
			ProjectID: ws.BkSystems.BkBCSProjectID,
			// 历史兼容 quirk：BCS.ProjectName 取 bkCIProjectID 而非 bcsProjectID
			ProjectName: ws.BkSystems.BkCIProjectID,
		}
	}

	// 填充 BKRepo
	if ws.BkSystems.BkRepoProjectID != "" {
		data.BKRepo = &bkiam.BKRepoOptions{
			ProjectID:   ws.BkSystems.BkRepoProjectID,
			ProjectName: ws.BkSystems.BkRepoProjectID,
		}
	}

	return data
}
