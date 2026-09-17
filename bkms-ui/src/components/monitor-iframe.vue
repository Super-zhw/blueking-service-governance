<!--
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
-->

<template>
  <Loading
    class="w-full h-full"
    :loading="isLoading"
  >
    <iframe
      ref="iframeRef"
      allow="fullscreen"
      allowfullscreen
      frameborder="0"
      height="100%"
      :src="url"
      width="100%"
      @error="isLoading = false"
      @load="handleIframeLoad"
    >
    </iframe>
  </Loading>
</template>
<script lang="ts" setup>
  import { onBeforeUnmount, onMounted, ref } from 'vue';

  import { Loading } from 'bkui-vue';
  import {
    type MonitorIframeRouteChangeMessage,
    type MonitorIframeSetParamsMessage,
    type MonitorIframeSetParamsPayload,
    BK_MONITOR_SOURCE,
    getMonitorOrigin,
    MONITOR_IFRAME_SOURCE,
  } from '~/composables/use-monitor-iframe';

  interface IProps {
    /** iframe URL（由父页面通过 useMonitorIframe().buildIframeUrl() 构建后传入，仅在初始化时计算） */
    url: string;
  }

  defineProps<IProps>();
  const emit = defineEmits<{
    (e: 'route-change', payload: { hash: string; href: string; query: Record<string, unknown> }): void;
  }>();

  const isLoading = ref(true);
  const iframeRef = ref<HTMLIFrameElement>();
  /**
   * iframe 文档加载完成标记：为 true 后 sendSetParams 才实时投递；未 true 时参数走 pending 补发
   * 前置约束：url prop 变化触发 src 重载时本标记不会自动复位，调用方须通过 v-if 重建组件或走 buildIframeUrl 冷启动，
   * 保证「url 更新与实例重建同生共死」，本组件不承诺同实例改 url 后的就绪语义
   */
  const isIframeReady = ref(false);
  /** iframe 未就绪期间收到的待同步参数：暂存后在 load 完成时补发，避免「iframe 重建/加载中切换环境」时参数静默丢失 */
  let pendingPayload: MonitorIframeSetParamsPayload | null = null;

  /** 监控平台 origin（postMessage targetOrigin 与 message 校验用），复用 hook 的纯函数避免重复实现 */
  const monitorOrigin = getMonitorOrigin();

  /** 构建 set-params 消息（发送与补发共用，保证消息结构单一来源） */
  function createSetParamsMessage(payload: MonitorIframeSetParamsPayload): MonitorIframeSetParamsMessage {
    return {
      source: MONITOR_IFRAME_SOURCE,
      type: 'set-params',
      payload,
    };
  }

  /** iframe 加载完成：标记就绪，并补发加载期间暂存的待同步参数（若存在） */
  function handleIframeLoad() {
    isLoading.value = false;
    isIframeReady.value = true;
    if (pendingPayload) {
      iframeRef.value?.contentWindow?.postMessage(createSetParamsMessage(pendingPayload), monitorOrigin);
      pendingPayload = null;
    }
  }

  /** 接收监控平台 route-change 消息（子 → 父） */
  function handleWindowMessage(event: MessageEvent) {
    // 仅接受本组件 iframe 发出的消息：规避同页多实例/来源不明消息导致 route-change 相互覆盖
    if (event.source !== iframeRef.value?.contentWindow) return;
    if (event.origin !== monitorOrigin) return;
    const data = event.data as MonitorIframeRouteChangeMessage | undefined;
    if (data?.source !== BK_MONITOR_SOURCE || data.type !== 'route-change') return;
    emit('route-change', {
      href: data.href,
      hash: data.hash,
      query: data.query,
    });
  }

  /** 向监控平台下发 filter-app_name / filter-service_name（父 → 子；set-params 支持双过滤条件）
   *  iframe 已就绪时直接发送；加载中/重建期先暂存 payload，待 load 完成后自动补发（此时返回 false，调用方无需额外处理） */
  function sendSetParams(payload: MonitorIframeSetParamsPayload): boolean {
    if (!monitorOrigin) return false;
    if (!isIframeReady.value) {
      pendingPayload = payload;
      return false;
    }
    iframeRef.value?.contentWindow?.postMessage(createSetParamsMessage(payload), monitorOrigin);
    return true;
  }

  onMounted(() => {
    window.addEventListener('message', handleWindowMessage);
  });

  onBeforeUnmount(() => {
    window.removeEventListener('message', handleWindowMessage);
  });

  defineExpose({ sendSetParams });
</script>
