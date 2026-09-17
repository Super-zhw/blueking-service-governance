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

// pina 持久化存储
import { get, has, merge, pickBy, set } from 'lodash-es';
import { createPinia } from 'pinia';
import { STORAGE_KEY, STORAGE_VERSION } from '~/common/const';

import type { PiniaPluginContext, Store } from 'pinia';
import type { UserModule } from '~/types.ts';

// 本地缓存配置信息
interface Options {
  ids: string[];
  key: string;
  paths: string[];
  storage: Storage;
  version: string;
  assertStorage?: (storage: Storage) => Error | void;
  getState: (key: string, storage: Storage) => any;
  reducer: (state: Store, paths: string[]) => object;
  setState: (key: string, state: object, storage: Storage) => void;
}
// 缓存接口定义
interface Storage {
  clear: () => void;
  getItem: (key: string) => any;
  removeItem: (key: string) => void;
  setItem: (key: string, value: any) => void;
}

// 校验缓存是否可用
function assertStorage(storage: Storage = window.localStorage) {
  try {
    storage.setItem('@@', 1);
    storage.removeItem('@@');
  } catch (err) {
    console.warn('Storage is not available:', err);
  }
}
// 获取缓存数据
function getState(key = '_pina_', storage: Storage = window.localStorage) {
  const value = storage.getItem(key);

  try {
    return value ? JSON.parse(value) : value;
  } catch (err) {}

  return undefined;
}
function installPiniaStorage(opt: Partial<Options> = {}) {
  const options: Options = merge(
    {
      key: '_pina_',
      version: '',
      overwrite: false,
      storage: window.localStorage,
      paths: [],
      ids: [],
      reducer,
      getState,
      setState,
      assertStorage,
    },
    opt,
  );

  assertStorage(options.storage);
  const savedState = options.getState(options.key, options.storage);

  return ({ store }: PiniaPluginContext) => {
    if (!options.ids?.includes(store.$id)) return;
    // 还原localstorage里面的值
    if (typeof savedState === 'object' && savedState !== null) {
      Object.keys(savedState).forEach(key => {
        if (has(store, key)) {
          set(store, key, savedState[key]);
        }
      });
    }

    // store 变化：仅用本 store 的非 undefined path 覆盖，与现有缓存合并写入。
    // 否则多 store 共享 paths 整体覆盖时，任一 store 变化都会把其他 store 的 path 冲掉
    // （如 space 变化冲掉 currentEnv、deploy-env 变化冲掉 statusTab）。
    // 注意：仅支持顶层 path 的浅合并，嵌套 path（如 preferences.foo）会整体覆盖，请勿注册。
    store.$subscribe(() => {
      const raw = options.getState(options.key, options.storage);
      const saved = raw && typeof raw === 'object' ? raw : {};
      const patch = pickBy(options.reducer(store, options.paths), value => value !== undefined);
      options.setState(options.key, { ...saved, ...patch }, options.storage);
    });
  };
}
function reducer(state: Store, paths: string[]) {
  return Array.isArray(paths) ? paths.reduce((substate, path) => set(substate, path, get(state, path)), {}) : state;
}

// 设置缓存状态
function setState(key = '_pina_', state: object, storage: Storage) {
  try {
    return storage.setItem(key, JSON.stringify(state));
  } catch (err) {
    console.warn('Failed to save state to storage:', err);
  }
}

// Setup Pinia
// https://pinia.vuejs.org/
export const install: UserModule = ({ app }) => {
  const pinia = createPinia().use(
    installPiniaStorage({
      key: STORAGE_KEY,
      version: STORAGE_VERSION,
      ids: ['user', 'space', 'deploy-env', 'ui'],
      paths: ['statusTab', 'lastAppTemplateID', 'currentEnv', 'preferences'],
    }),
  );
  app.use(pinia);
};
