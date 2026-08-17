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

// Package underlayip provides the underlay-ip section subcommand.
package underlayip

import "github.com/spf13/cobra"

// NewCmd creates the underlay-ip section command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "underlay-ip",
		Short: "Manage underlay IP (VPC-CNI) networking",
		Long: `Manage the underlay IP networking mode for the application.

When enabled, pods are assigned a VPC IP address directly, allowing cross-VPC network
access. This section supports both default (application-level) and environment-level
configuration.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewViewCmd())
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewDisableCmd())
	cmd.AddCommand(NewResetCmd())

	return cmd
}
