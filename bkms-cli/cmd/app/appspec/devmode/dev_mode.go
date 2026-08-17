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

// Package devmode provides the dev-mode section subcommand.
package devmode

import "github.com/spf13/cobra"

// NewCmd creates the dev-mode section command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev-mode",
		Short: "Manage development mode (env-level only)",
		Long: `Manage the development mode for an application environment.

Development mode allows developers to hot-update application binaries at runtime without
rebuilding or redeploying the container image, enabling rapid iteration during development.
Once enabled, use 'bkms-cli app publish' to push new binaries to target instances.
See 'bkms-cli app publish --help' for details.

This is an environment-level only configuration (not available in production), so --env
is required for every subcommand.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewViewCmd())
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewDisableCmd())
	cmd.AddCommand(NewResetCmd())

	return cmd
}
