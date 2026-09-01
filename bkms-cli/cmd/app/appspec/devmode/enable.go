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

package devmode

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
	apphandler "github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/handler/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/handler/appspec"
	cmdutil "github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/cmd"
)

// NewEnableCmd returns a Command instance for 'appspec dev-mode enable' sub command.
func NewEnableCmd() *cobra.Command {
	var appID, workspaceID, envName string

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable dev-mode for an environment",
		Long: `Enable the development mode for an application environment.

After enabling, the application will need to be redeployed to take effect. Note that
production environments do not support development mode.`,
		Example: `  # Enable dev-mode for an environment
  bkms-cli app appspec dev-mode enable --app my-app --env prod`,
		PreRun: cmdutil.CommonPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedAppID, err := apphandler.ResolveAppID(
				cmd.Context(), cmdutil.GetWorkspaceID(workspaceID), appID,
			)
			if err != nil {
				return errors.Wrap(err, "resolve app")
			}
			appID = resolvedAppID

			if err := appspec.SetEnabledHandler(
				cmd.Context(), appID, envName, client.AppSpecSectionDevMode, true,
			); err != nil {
				return errors.Wrap(err, "enable dev-mode")
			}

			fmt.Printf("Successfully enabled dev-mode for app %s in env %s\n", appID, envName)
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app", "", "application ID or name (required)")
	cmd.Flags().StringVar(&workspaceID, "workspace", "", "workspace ID")
	cmd.Flags().StringVar(&envName, "env", "", "environment name (required)")

	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("env")

	return cmd
}
