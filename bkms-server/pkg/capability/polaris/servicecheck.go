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

// Package polaris 声明与框架相关的北极星能力：服务名校验与注册信息注入。北极星注册本身不在此声明。
package polaris

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app/appcfg"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

// ServiceCheckProps 无参数。
type ServiceCheckProps struct{}

// Public 无对外内容。
func (p ServiceCheckProps) Public() any { return nil }

// ServiceCheck 北极星服务名校验：北极星配置的服务名不在 tRPC 配置文件中时给出警告。仅 tRPC 支持。
var ServiceCheck = runtimemod.Spec[ServiceCheckProps]{
	ID: runtimemod.ModPolarisServiceCheck,
	Resolve: runtimemod.ByFramework(map[string]ServiceCheckProps{
		appruntime.FrameworkTRPC: {},
	}),
}

// ServiceCheckTarget 待校验的北极星配置。
type ServiceCheckTarget struct {
	// AppID 应用 ID。
	AppID string
	// ConfigName 北极星配置名，用于警告文案。
	ConfigName string
	// PolarisName 北极星服务名。
	PolarisName string
	// EnvNames 逐个校验的环境。
	EnvNames []string
}

// CheckServiceName 逐个环境校验北极星服务名是否出现在 tRPC 配置的服务名中，返回不一致或读取失败的警告。
// 应用不支持该能力时不校验。
func CheckServiceName(
	ctx context.Context,
	provider appcfg.MountableFileProvider,
	mod runtimemod.Mod[ServiceCheckProps],
	target ServiceCheckTarget,
) (warnings []string) {
	if !mod.Supported {
		return nil
	}
	for _, envName := range target.EnvNames {
		if warning := checkServiceNameInEnv(ctx, provider, target, envName); warning != "" {
			warnings = append(warnings, warning)
		}
	}
	return warnings
}

// checkServiceNameInEnv 校验单个环境，一致时返回空字符串
func checkServiceNameInEnv(
	ctx context.Context,
	provider appcfg.MountableFileProvider,
	target ServiceCheckTarget,
	envName string,
) string {
	serviceNames, err := getTrpcServiceNames(ctx, provider, target.AppID, envName)
	if err != nil {
		if envName == "" {
			return fmt.Sprintf("[%s] 读取 tRPC 配置文件失败: %v", target.ConfigName, err)
		}
		return fmt.Sprintf("[%s] 环境 '%s': 读取 tRPC 配置文件失败: %v", target.ConfigName, envName, err)
	}

	if !lo.Contains(serviceNames, target.PolarisName) {
		if envName == "" {
			return fmt.Sprintf("[%s] 北极星服务名 '%s' 推荐与 tRPC 配置中的服务名 %v 一致",
				target.ConfigName, target.PolarisName, serviceNames)
		}
		return fmt.Sprintf("[%s] 环境 '%s': 北极星服务名 '%s' 推荐与 tRPC 配置中的服务名 %v 一致",
			target.ConfigName, envName, target.PolarisName, serviceNames)
	}
	return ""
}

// trpcServiceConfig 是用于提取 tRPC 配置中 server.service 信息的轻量结构体
// 当前业务只需要解析 server.service[].name，因此这里只覆盖少量字段；完整字段可参考：
// https://github.com/trpc-group/trpc-go/blob/9b5c63e5/config.go#L514
type trpcServiceConfig struct {
	Server struct {
		Service []*struct {
			Name string `yaml:"name"`
		} `yaml:"service"`
	} `yaml:"server"`
}

// getTrpcServiceNames 获取指定应用和环境下 tRPC 配置文件中的所有服务名
func getTrpcServiceNames(
	ctx context.Context,
	provider appcfg.MountableFileProvider,
	appID, envName string,
) ([]string, error) {
	frameworkFile, err := provider.GetFrameworkMountableFile(ctx, appID, envName)
	if err != nil {
		return nil, err
	}
	return parseTrpcServiceNames(frameworkFile.Content)
}

// parseTrpcServiceNames 从 tRPC 配置 YAML 内容中提取所有 server.service[].name
func parseTrpcServiceNames(content string) ([]string, error) {
	var cfg trpcServiceConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshal tRPC config YAML")
	}

	var names []string
	for _, svc := range cfg.Server.Service {
		if svc != nil && svc.Name != "" {
			names = append(names, svc.Name)
		}
	}
	return names, nil
}
