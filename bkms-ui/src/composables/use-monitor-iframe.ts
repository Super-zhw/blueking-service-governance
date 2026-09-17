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
import { ref, watch } from 'vue';

import { ApiServerService } from '~/api/modules/bkmsserver';
import { objectToUrlParams } from '~/common/util';
import { useSpaceStore } from '~/stores/space';

import type { ApmQueryParams } from '~/common/const';

/** iframe 前缀公共参数 */
export interface MonitorIframeBaseQueryParams {
  apm_nav_list?: boolean;
  apm_submenu?: number;
  needMenu?: boolean | number;
}

/** 观测参数（父页面传入形态：filter-app_name 必填，其余字段可缺省由页面默认值补全） */
export type MonitorIframeObservabilityType = Partial<Omit<ApmQueryParams, 'filter-app_name'>> &
  Required<Pick<ApmQueryParams, 'filter-app_name'>>;

/** 监控平台 → 父页面：路由变化通知 */
export interface MonitorIframeRouteChangeMessage {
  hash: string;
  href: string;
  query: Record<string, unknown>;
  source: string;
  type: 'route-change';
}

/** 父页面 → 监控平台：下发参数 */
export interface MonitorIframeSetParamsMessage {
  payload: MonitorIframeSetParamsPayload;
  source: string;
  type: 'set-params';
}

/** 下发参数：监控平台 set-params 支持 app_name 与 service_name 双过滤条件，二者均必填 */
export type MonitorIframeSetParamsPayload = {
  'filter-app_name': string;
  'filter-service_name': string;
};

/** 我方消息标识（与监控平台约定） */
export const MONITOR_IFRAME_SOURCE = 'bk-service-governance';
/** 监控平台消息标识 */
export const BK_MONITOR_SOURCE = 'bk-monitor';

/** 监控平台 origin（postMessage targetOrigin 与 message 校验用），URL 解析规则变更时仅需改此处 */
export function getMonitorOrigin(): string {
  try {
    return new URL(import.meta.env.BK_MONITOR).origin;
  } catch {
    return '';
  }
}

/**
 * monitor-iframe 公共能力：bizId 获取、iframe URL 构建（监控平台 origin 为独立纯函数 getMonitorOrigin，调用方按需直接 import）
 *
 * 抽成 hook 的原因：
 * 1. URL 构建依赖的公共属性（bizId、baseQueryParams、类型、协议常量）在 observation / env-detail 两个页面一致，
 *    抽离后保证「类型 + 公共属性」单一来源，避免两处各自维护导致漂移
 * 2. iframe 初始化后 URL 锁定，后续参数变更仅走 postMessage——URL 构建时机（bizId 就绪）封装在 hook 内，
 *    页面侧只需在观测参数就绪后调用 buildIframeUrl，职责边界清晰
 */
export function useMonitorIframe(type: 'application' | 'service', baseQueryParams: MonitorIframeBaseQueryParams) {
  const spaceStore = useSpaceStore();

  /** 监控平台业务 ID（bkMonitorProjectID），随空间切换自动获取 */
  const bizId = ref<string>('');
  /** fetchBizId 进行中/已完成 的 Promise，用于去重：watch 与 buildIframeUrl 兜底并发调用时不重复请求 */
  let bizIdPromise: null | Promise<void> = null;
  /** fetchBizId 请求序号：空间切换作废旧缓存后，旧请求晚返回时丢弃，防止旧空间数据覆盖新空间 bizId */
  let fetchBizIdSeq = 0;

  function fetchBizId() {
    if (!bizIdPromise) {
      const seq = ++fetchBizIdSeq;
      bizIdPromise = ApiServerService.GetWorkspace({
        workspaceID: spaceStore.currentSpace,
      })
        .then(workspaceData => {
          // 仅接受最新一次请求的结果：空间切换后旧请求晚返回不覆盖新空间 bizId
          if (seq === fetchBizIdSeq && workspaceData) {
            bizId.value = workspaceData.bkSystems?.bkMonitorProjectID || '';
          }
        })
        .catch(() => undefined);
    }
    return bizIdPromise;
  }

  /**
   * 构建 iframe URL：拼接前缀（needMenu/bizId/parentOrigin）+ hash 路由（观测参数 + sceneId）
   * 调用时机：observabilityQuery 就绪（filter-app_name 等字段完整）后调用；bizId 通常由内部 watch 提前获取，此处兜底再取一次（fetchBizId 已去重）
   */
  async function buildIframeUrl(observabilityQuery: MonitorIframeObservabilityType) {
    if (!bizId.value) await fetchBizId();
    const prefixParams = objectToUrlParams({
      ...baseQueryParams,
      bizId: bizId.value,
      parentOrigin: window.location.origin,
    });
    const suffixParams = objectToUrlParams({
      ...observabilityQuery,
      sceneId: type === 'service' ? 'apm_service' : 'apm_application',
    });
    return `${import.meta.env.BK_MONITOR}/?${prefixParams}#/apm/${type}?${suffixParams}`;
  }

  // 空间变化（含首次进入）时自动拉取 bizId，调用方无需关心获取时机
  // 切换空间会作废正在进行的请求缓存并重新拉取：正常路由下页面随 App.vue RouterView key 重建、hook 实例随之重建，
  // 此处仅为 keep-alive 等「同实例跨空间」场景作废 bizId 缓存，保证后续 buildIframeUrl 取到新空间 bizId；
  // 注意：页面观测参数/iframeUrl 的跨空间重建仍需另行支持，此兜底不覆盖该链路
  watch(
    () => spaceStore.currentSpace,
    () => {
      bizIdPromise = null;
      bizId.value = '';
      fetchBizId();
    },
    { immediate: true },
  );

  return {
    buildIframeUrl,
  };
}
