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

package deploy

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
	deployhandler "github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/handler/deploy"
	cmdutil "github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/cmd"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/console"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/output"
)

// NewPrecheckCmd returns a Command instance for 'app deploy precheck' sub command
func NewPrecheckCmd() *cobra.Command {
	var appID, envName, outputFormat string

	cmd := &cobra.Command{
		Use:   "precheck",
		Short: "Check if all environment variables are defined before deployment",
		Long: `Pre-deployment check: verifies that all environment variables referenced in
config files and start commands are defined.

Exit code 0 means all variables are defined (safe to deploy).
Exit code 1 means there are undefined variables that must be resolved first.

Only supported for trpc and taf application types.`,
		Example: `  # Check before deploying to prod
  bkms-cli app deploy precheck --app myapp --env prod

  # Output as JSON for scripting
  bkms-cli app deploy precheck --app myapp --env prod -o json`,
		PreRunE: cmdutil.ResolveAppPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeployPrecheck(cmd, appID, envName, outputFormat)
		},
	}

	cmdutil.AddAppFlags(cmd, &appID)
	cmd.Flags().StringVar(&envName, "env", "", "environment name (required)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", output.FlagUsage)

	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("env")

	return cmd
}

func runDeployPrecheck(cmd *cobra.Command, appID, envName, outputFormat string) error {
	result, err := deployhandler.PrecheckEnvVars(cmd.Context(), appID, envName)
	if err != nil {
		return errors.Wrap(err, "precheck deploy env vars")
	}

	if outputFormat != "" {
		formatted, fmtErr := output.FormatData(cmd.Context(), result, outputFormat)
		if fmtErr != nil {
			return errors.Wrap(fmtErr, "format output")
		}
		console.Info("%s", formatted)
		if !result.Passed {
			return errors.New("precheck failed")
		}
		return nil
	}

	if result.Passed {
		console.Info("✓ Pre-check passed: all environment variables are defined for app %s in env %s",
			appID, envName)
		return nil
	}

	printPrecheckFailure(result)
	return errors.New("precheck failed: undefined environment variables found")
}

func printPrecheckFailure(result *client.DeployPrecheckResult) {
	console.Warn("✗ Pre-check FAILED: %d undefined environment variable(s) found", len(result.UndefinedVars))
	console.Info("")
	console.Info("  %-30s %s", "KEY", "REFERENCED BY")
	console.Info("  %-30s %s", strings.Repeat("-", 30), strings.Repeat("-", 40))
	for _, v := range result.UndefinedVars {
		refs := lo.Map(v.Sources, func(s client.EnvVarSource, _ int) string {
			if s.Name != "" {
				return fmt.Sprintf("%s:%s", s.Type, s.Name)
			}
			return s.Type
		})
		console.Info("  %-30s %s", v.Key, strings.Join(refs, ", "))
	}
	console.Info("")
	console.Tips("Fix: use 'bkms-cli envvar create' to define missing variables, then re-run precheck.")
}
