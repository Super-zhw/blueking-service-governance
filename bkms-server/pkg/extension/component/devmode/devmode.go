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

// Package devmode 提供开发模式组件支持
//
// 开发模式允许开发者在容器运行时动态替换二进制文件并重启服务，无需重新构建和部署镜像。
// 支持 development、test 和 staging 环境，仅 production 环境严禁使用。
// 支持 trpc 和 taf 两种应用类型，两者使用不同的工作目录和脚本集。
//
// 工作原理：将管理脚本打包为 ConfigMap，挂载到容器后替换启动命令为 init.sh，
// 由 monitor.sh 守护进程负责进程监控和自动重启。
package devmode

import (
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	devmodecap "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/capability/devmode"
	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod/mods"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/appmodel"
)

// DevMode 开发模式组件接口
type DevMode interface {
	// IsAllowed 检查当前环境是否允许使用开发模式
	IsAllowed() bool

	// Validate 验证配置有效性
	Validate() error

	// Build 构建完整输出
	Build() (*Output, error)

	// BuildConfigMap 构建脚本 ConfigMap
	BuildConfigMap() (*corev1.ConfigMap, error)

	// BuildVolume 构建 Volume
	BuildVolume() corev1.Volume

	// BuildVolumeMount 构建 VolumeMount
	BuildVolumeMount() corev1.VolumeMount

	// BuildCommand 构建启动命令
	BuildCommand() []string
}

// Config 开发模式配置
type Config struct {
	// Enabled 是否启用
	Enabled bool

	// EnvType 环境类型
	EnvType string

	// Props 平台为该技术栈声明的脚本集与路径，为空表示不支持开发模式
	Props devmodecap.Props

	// AppName 应用名称，应与二进制文件名一致
	AppName string

	// StartupCommand 用户 app 启动命令
	StartupCommand string

	// StartupShCommand 用户 app 启动脚本，非空时写入独立脚本文件以 /bin/sh 执行（$0 为脚本路径）
	StartupShCommand string
}

// IsTaf 是否使用 TAF 脚本集
func (c *Config) IsTaf() bool {
	return c.Props.ScriptSet == devmodecap.ScriptSetTAF
}

// IsTrpc 是否使用 tRPC 脚本集
func (c *Config) IsTrpc() bool {
	return c.Props.ScriptSet == devmodecap.ScriptSetTRPC
}

// Output 开发模式组件输出
type Output struct {
	// ConfigMap 包含所有管理脚本的 ConfigMap
	ConfigMap *corev1.ConfigMap

	// Volume 引用 ConfigMap 的 Volume
	Volume corev1.Volume

	// VolumeMount 挂载到容器的 VolumeMount
	VolumeMount corev1.VolumeMount

	// Command 容器启动命令（指向 init.sh）
	Command []string
}

// RequireProps 返回应用的开发模式能力参数。应用的技术栈不支持开发模式时返回
// ErrUnsupportedAppType。开发模式的各个入口（配置写入、预检、重启、渲染）统一通过它判断。
func RequireProps(app *bkmsapp.Application) (devmodecap.Props, error) {
	props, err := mods.Of(app).DevMode.Require()
	if err != nil {
		return devmodecap.Props{}, errors.Wrapf(ErrUnsupportedAppType, "app %s: %v", app.ID, err)
	}
	return props, nil
}

// New 创建开发模式构建器实例
func New(config *Config) DevMode {
	if config == nil {
		return &builder{config: &Config{Enabled: false}}
	}
	return &builder{config: config}
}

// CreateDevModeConfig 从 AppModel、能力参数和环境信息创建开发模式配置
func CreateDevModeConfig(
	appModel *appmodel.AppModel, props devmodecap.Props, envType string, enabled bool,
) *Config {
	config := FromAppModel(appModel, props, envType)
	config.Enabled = enabled
	return config
}

// FromAppModel 从 AppModel 与能力参数创建开发模式配置
func FromAppModel(appModel *appmodel.AppModel, props devmodecap.Props, envType string) *Config {
	// 从 appModel.Workload.Command、Args 中提取
	var customStartScript string

	// 检查是否有用户 app 启动命令
	if len(appModel.Workload.Command) > 0 || len(appModel.Workload.Args) > 0 {
		temp := make([]string, 0)
		temp = append(temp, appModel.Workload.Command...)
		temp = append(temp, appModel.Workload.Args...)
		customStartScript = strings.Join(temp, " ")
	}

	return &Config{
		Enabled:          false,
		EnvType:          envType,
		Props:            props,
		AppName:          appModel.Workload.Name,
		StartupCommand:   customStartScript,
		StartupShCommand: appModel.Workload.ShCommand,
	}
}

// ConfigMapResourceName 根据应用名称生成开发模式资源名称（ConfigMap / Volume）
// 格式: {appName}-devmode-scripts
func ConfigMapResourceName(appName string) string {
	return appName + ConfigMapResourceSuffix
}
