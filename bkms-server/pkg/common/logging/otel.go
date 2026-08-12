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

	"github.com/go-slog/otelslog"
	slogmulti "github.com/samber/slog-multi"
	otelslogbridge "go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// InitOTelHandler 注入 OTel Handler 到全局 slog，构建 Pipe(traceMiddleware) → Fanout(local, otelBridge) pipeline。
// provider 为 nil 时仅注入 trace_id/span_id 中间件，不上报远程。
func InitOTelHandler(provider *sdklog.LoggerProvider) {
	localHandler := slog.Default().Handler()

	if provider == nil {
		slog.SetDefault(slog.New(
			slogmulti.Pipe(otelslog.Middleware()).Handler(localHandler),
		))
		return
	}

	otelHandler := otelslogbridge.NewHandler("bkms", otelslogbridge.WithLoggerProvider(provider))

	slog.SetDefault(slog.New(
		slogmulti.Pipe(otelslog.Middleware()).Handler(
			slogmulti.Fanout(localHandler, otelHandler),
		),
	))
}
