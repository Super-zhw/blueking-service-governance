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

package logging

import (
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
	otelslogbridge "go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// InitOTelHandler 追加 OTel bridge handler 到全局 slog，实现日志远程上报。
// 调用前 InitDefaultLogger 已挂载 otelslog.Middleware()（trace_id/span_id 注入），
// 此处仅通过 Fanout 追加 otelBridge handler，不重复包装 Middleware。
func InitOTelHandler(provider *sdklog.LoggerProvider) {
	localHandler := slog.Default().Handler()
	otelHandler := otelslogbridge.NewHandler("bkms", otelslogbridge.WithLoggerProvider(provider))

	slog.SetDefault(slog.New(
		slogmulti.Fanout(localHandler, otelHandler),
	))
}
