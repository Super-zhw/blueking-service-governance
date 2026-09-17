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

import { ref } from 'vue';

import { defineStore } from 'pinia';

// UI 偏好总集：key 为偏好标识（如 table:deploy-overview），value 为偏好数据。
// store 只做通用存取，具体偏好的类型由使用方（hook）在读写时收敛。
// 持久化由 modules/pinia.ts 的 installPiniaStorage 统一处理（paths 注册 'preferences'）。
export const useUiStore = defineStore('ui', () => {
  const preferences = ref<Record<string, unknown>>({});

  // 写入偏好；value 传 null/undefined 表示清除该偏好，避免留下无意义记录。
  function setPreference(key: string, value: unknown) {
    if (value === null || value === undefined) {
      const next = { ...preferences.value };
      delete next[key];
      preferences.value = next;
      return;
    }
    preferences.value = { ...preferences.value, [key]: value };
  }

  return { preferences, setPreference };
});
