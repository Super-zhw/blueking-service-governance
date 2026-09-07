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

package dashboard

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// appDashboardCollectionName is the MongoDB collection name for storing app dashboard bindings.
const appDashboardCollectionName = "bkmonitor_app_dashboard"

var (
	// ErrNotFound 应用仪表盘绑定未找到
	ErrNotFound = errors.New("AppDashboard not found")

	// ErrDuplicate 应用已绑定相同 uid 的仪表盘
	ErrDuplicate = errors.New("AppDashboard duplicate uid error")

	// ErrDashboardNotExist 提交的仪表盘 uid 在 bkmonitor 侧不存在
	ErrDashboardNotExist = errors.New("dashboard uid does not exist in bkmonitor")
)

var _ AppDashboardStore = &StoreMongo{}

// AppDashboardStore defines the storage interface for app dashboard binding data.
type AppDashboardStore interface {
	// Create creates a new app dashboard binding.
	Create(ctx context.Context, d *AppDashboard) (bson.ObjectID, error)

	// ListByApp returns app dashboard bindings by app ID.
	ListByApp(ctx context.Context, appID string) ([]AppDashboard, error)

	// Update updates an app dashboard binding by app ID and uid.
	Update(ctx context.Context, appID, uid string, updateData *AppDashboardUpdateData) error

	// Delete deletes an app dashboard binding by app ID and uid.
	Delete(ctx context.Context, appID, uid string) error
}

// StoreMongo implements AppDashboardStore interface with MongoDB.
type StoreMongo struct {
	collection *mongo.Collection
}

// NewStoreMongo creates a new AppDashboardStore instance.
func NewStoreMongo(client *mongo.Client, dbName string) (AppDashboardStore, error) {
	coll := client.Database(dbName).Collection(appDashboardCollectionName)
	// 索引（由 golang-migrate 维护）：
	// - 唯一：appID + uid
	return &StoreMongo{collection: coll}, nil
}

// Create creates a new app dashboard binding.
func (s *StoreMongo) Create(ctx context.Context, d *AppDashboard) (bson.ObjectID, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(d); err != nil {
		return bson.NilObjectID, errors.Wrap(err, "AppDashboard validation failed")
	}

	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	d.UpdatedAt = d.CreatedAt

	ret, err := s.collection.InsertOne(ctx, d)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return bson.NilObjectID, ErrDuplicate
		}
		return bson.NilObjectID, err
	}
	return ret.InsertedID.(bson.ObjectID), nil
}

// ListByApp returns app dashboard bindings by app ID.
func (s *StoreMongo) ListByApp(ctx context.Context, appID string) ([]AppDashboard, error) {
	filter := bson.M{"appID": appID}
	sort := bson.D{{Key: "createdAt", Value: -1}}
	findOptions := options.Find().SetSort(sort)

	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx) // nolint

	list := make([]AppDashboard, 0)
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// Update updates an app dashboard binding by app ID and uid.
func (s *StoreMongo) Update(
	ctx context.Context,
	appID, uid string,
	updateData *AppDashboardUpdateData,
) error {
	if updateData == nil {
		return nil
	}

	update, isEmpty := updateData.ToBSON()
	if !isEmpty {
		update["updatedAt"] = time.Now()
		return s.updateOne(ctx, bson.M{"appID": appID, "uid": uid}, bson.M{"$set": update})
	}
	return nil
}

// Delete deletes an app dashboard binding by app ID and uid.
func (s *StoreMongo) Delete(ctx context.Context, appID, uid string) error {
	ret, err := s.collection.DeleteOne(ctx, bson.M{"appID": appID, "uid": uid})
	if err != nil {
		return err
	}
	if ret.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *StoreMongo) updateOne(ctx context.Context, filter, update bson.M) error {
	opts := options.UpdateOne().SetUpsert(false)
	ret, err := s.collection.UpdateOne(ctx, filter, update, opts)
	if ret != nil && ret.MatchedCount == 0 {
		return ErrNotFound
	}
	return err
}
