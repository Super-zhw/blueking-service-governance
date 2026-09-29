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

package workspace

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/account/auth"
	bscpapi "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/cloudapi/bscp"
)

// bscpCredentialName 固定的 Credential 名称，每个 BSCP 项目下只有一个
const bscpCredentialName = "bkms-credential" // nolint: gosec

// BscpBinding 绑定 BSCP 项目的结果，用于写入 workspace 的 BkSystems。
type BscpBinding struct {
	// ProjectID BSCP 项目 ID
	ProjectID string
	// ProjectKey BSCP 项目 Key
	ProjectKey string
	// CredentialID BSCP Credential ID
	CredentialID string
	// Token BSCP Credential 的访问令牌（用于 sidecar 拉取配置）
	Token string
}

// BindBscpProject 绑定 workspace 到 BSCP 项目：解析项目（显式 projectKey 或 Default 项目），
// 并确保 credential 存在，返回需要写入 workspace 的字段。
func BindBscpProject(ctx context.Context, bizID, projectKey string) (*BscpBinding, error) {
	client, err := bscpapi.NewConfigClient(auth.MustGetUser(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "create bscp config client")
	}

	project, err := resolveProject(ctx, client, bizID, projectKey)
	if err != nil {
		return nil, err
	}

	credentialID, token, err := ensureBscpCredential(ctx, client, bizID, project.ID)
	if err != nil {
		return nil, err
	}

	return &BscpBinding{
		ProjectID:    cast.ToString(project.ID),
		ProjectKey:   project.Spec.Key,
		CredentialID: credentialID,
		Token:        token,
	}, nil
}

// resolveProject 解析要绑定的 BSCP 项目：显式指定 projectKey 时按 key 查询，否则选择 Default 项目。
func resolveProject(
	ctx context.Context,
	client bscpapi.ConfigClient,
	bizID, projectKey string,
) (*bscpapi.Project, error) {
	if projectKey != "" {
		project, err := client.GetProjectByKey(ctx, bizID, projectKey)
		if err != nil {
			return nil, errors.Wrapf(err, "get bscp project by key %s", projectKey)
		}
		return project, nil
	}

	projects, err := client.ListProjects(ctx, bizID)
	if err != nil {
		return nil, errors.Wrapf(err, "list bscp projects for biz %s", bizID)
	}
	project := bscpapi.DefaultProject(projects)
	if project == nil {
		return nil, errors.Errorf("no bscp project found under biz %s", bizID)
	}
	return project, nil
}

// ensureBscpCredential 确保 BSCP 项目的 credential 存在（幂等），返回 credential ID 和 token。
func ensureBscpCredential(
	ctx context.Context,
	client bscpapi.ConfigClient,
	bizID string,
	projectID int64,
) (string, string, error) {
	credentials, err := client.ListCredentials(ctx, bizID, projectID)
	if err != nil {
		return "", "", errors.Wrapf(err, "list credentials for biz %s, project %d", bizID, projectID)
	}
	for _, c := range credentials {
		if c.Name == bscpCredentialName {
			return cast.ToString(c.ID), c.EncCredential, nil
		}
	}

	if _, err = client.CreateCredential(ctx, &bscpapi.CreateCredentialReq{
		BizID:     bizID,
		ProjectID: projectID,
		Name:      bscpCredentialName,
		Memo:      "auto-created by bkms platform",
	}); err != nil {
		return "", "", errors.Wrapf(
			err, "create credential %q in biz %s, project %d", bscpCredentialName, bizID, projectID,
		)
	}

	credentials, err = client.ListCredentials(ctx, bizID, projectID)
	if err != nil {
		return "", "", errors.Wrapf(
			err, "list credentials after creation for biz %s, project %d", bizID, projectID,
		)
	}
	for _, c := range credentials {
		if c.Name == bscpCredentialName {
			return cast.ToString(c.ID), c.EncCredential, nil
		}
	}
	return "", "", errors.Errorf(
		"credential %q not found after creation in biz %s, project %d", bscpCredentialName, bizID, projectID,
	)
}
