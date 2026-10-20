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
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/bkerrs"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/config"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/account/auth"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/cloudapi/bcs"
)

// FetchActiveToken 以当前登录用户身份获取 BCS 项目下状态为 active 的 Token。
// 同步接口与异步任务都用它获取用户 Token；任务在 worker 中调用，可避免把 Token 写入队列。
func FetchActiveToken(ctx context.Context, bcsProjectCode string) (string, error) {
	user, err := auth.GetUser(ctx)
	if err != nil {
		return "", bkerrs.Wrap(err, bkerrs.ErrCodeInternalServerError, "get auth user")
	}

	bcsClient, err := bcs.New(user, bcs.ClientOption{UserMode: true})
	if err != nil {
		return "", bkerrs.Wrap(err, bkerrs.ErrCodeInternalServerError, "failed to initialize token client")
	}
	tokens, err := bcsClient.ListUserTokens(ctx)
	if err != nil {
		return "", bkerrs.Wrap(err, bkerrs.ErrCodeInternalServerError, "failed to retrieve token")
	}

	activeToken, found := lo.Find(tokens, func(item bcs.UserToken) bool {
		return item.Status == bcs.UserTokenStatusActive
	})
	if !found {
		return "", bkerrs.Errorf(
			bkerrs.ErrCodeNotFound,
			"no active token found, please create a token in BCS at %s",
			buildBCSUserTokenURL(bcsProjectCode),
		)
	}
	return activeToken.Token, nil
}

// buildBCSUserTokenURL 构造 BCS 用户 Token 创建页地址
func buildBCSUserTokenURL(bcsProjectCode string) string {
	return fmt.Sprintf(
		"%s/bcs/projects/%s/user-token",
		strings.TrimRight(config.G.BCS.BaseUrl, "/"),
		bcsProjectCode,
	)
}
