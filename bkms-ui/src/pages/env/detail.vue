<!--
 - TencentBlueKing is pleased to support the open source community by making
 - 蓝鲸智云 - 服务治理 (BlueKing Service Governance) available.
 - Copyright (C) Tencent. All rights reserved.
 - Licensed under the MIT License (the "License"); you may not use this file except
 - in compliance with the License. You may obtain a copy of the License at
 -
 -  http://opensource.org/licenses/MIT
 -
 - Unless required by applicable law or agreed to in writing, software distributed under
 - the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 - either express or implied. See the License for the specific language governing permissions and
 - limitations under the License.
 -
 - We undertake not to change the open source license (MIT license) applicable
 - to the current version of the project delivered to anyone in the future.
-->

<template>
  <SlideDetail
    ref="slideDetailRef"
    class="bg-[#fff]"
  >
    <div class="h-[82px] pt-[16px] bg-[#f0f1f5] px-[24px]">
      <div class="flex items-center gap-[8px]">
        <div class="text-[14px] text-[#313238] font-bold">{{ data?.displayName }}</div>
        <Tag
          v-if="envTypeConfig"
          :class="envTypeTagClassMap[data?.type || '']"
          type="stroke"
          >{{ envTypeConfig.name || data?.type }}</Tag
        >
        <Dropdown
          ref="dropdownRef"
          placement="bottom-start"
          trigger="click"
        >
          <div
            class="w-[26px] h-[26px] border-[1px] border-[#C4C6CC] border-solid bg-[#FFF] rounded-[2px] flex justify-center items-center cursor-pointer"
          >
            <Ellipsis class="transform-rotate-90 text-[#4D4F56] text-[16px]" />
          </div>
          <template #content>
            <Dropdown.DropdownMenu>
              <Dropdown.DropdownItem @click="handleDeleteEnv">
                {{ $t('删除环境') }}
              </Dropdown.DropdownItem>
            </Dropdown.DropdownMenu>
          </template>
        </Dropdown>
      </div>
      <p class="mt-[6px] text-[12px] text-[#979BA5]">{{ data?.name }}</p>
    </div>
    <Tab
      :key="props?.data?.name"
      v-model:active="activeTabName"
      class="app-detail-tab"
      :label-height="42"
      type="card-tab"
      :validate-active="false"
    >
      <!-- <Tab.TabPanel name="overview" :label="$t('概览')">概览</Tab.TabPanel>
      <Tab.TabPanel name="applicationList" :label="$t('应用列表')">应用列表</Tab.TabPanel>
      <Tab.TabPanel name="laneSettings" :label="$t('泳道设置')">
        <div class="flex gap-[12px]">
          <Select>
            <Select.Option value="泳道A">{{ '泳道A' }}</Select.Option>
            <Select.Option value="泳道B">{{ '泳道B' }}</Select.Option>
          </Select>
          <Tab
            :label-height="42"
            :validate-active="false"
            type="card">
            <Tab.TabPanel name="details" :label="$t('详情')"></Tab.TabPanel>
            <Tab.TabPanel name="configuration" :label="$t('配置')"></Tab.TabPanel>
            <Tab.TabPanel name="observabilityData" :label="$t('观测数据')"></Tab.TabPanel>
            <Tab.TabPanel name="deploymentRecords" :label="$t('部署记录')"></Tab.TabPanel>
          </Tab>
        </div>
      </Tab.TabPanel>
      <Tab.TabPanel name="trafficManagement" :label="$t('流量治理')">流量治理</Tab.TabPanel>
      <Tab.TabPanel name="canaryStrategy" :label="$t('灰度策略')">灰度策略</Tab.TabPanel>
      <Tab.TabPanel name="setting" :label="$t('设置')">设置</Tab.TabPanel> -->
      <Tab.TabPanel
        :label="$t('基本信息')"
        name="basicInfo"
        render-directive="if"
      >
        <BasicInfo
          :key="tabKey"
          :env="data.id ?? ''"
          :workspace="workspace"
          @update="handleUpdate"
        />
      </Tab.TabPanel>
      <Tab.TabPanel
        :label="$t('环境配置')"
        name="setting"
        render-directive="if"
      >
        <Setting
          :key="tabKey"
          :env="data.id ?? ''"
          :env-display-name="data.displayName ?? data.name ?? ''"
          :env-name="data.name ?? ''"
          :env-type="data.type ?? ''"
          :workspace="workspace"
          @update="handleUpdate"
        />
      </Tab.TabPanel>
      <!-- <Tab.TabPanel
        :label="$t('泳道配置')"
        name="laneConfig"
        render-directive="if"
      >
        <LaneConfig :data="data" />
      </Tab.TabPanel> -->
      <Tab.TabPanel
        :label="$t('观测数据')"
        name="observability"
        render-directive="if"
      >
        <div class="flex flex-col gap-[12px] h-full">
          <ApmInstance
            :current-apm="currentApm"
            :data="data"
            @update:current-apm="getEnvApm"
          />
          <MonitorIframe
            v-if="currentApm"
            class="flex-1 min-h-0"
            :url="iframeUrl"
            @route-change="handleRouteChange"
          />
        </div>
      </Tab.TabPanel>
    </Tab>
  </SlideDetail>
</template>
<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';

  import { Dropdown, Tab, Tag } from 'bkui-vue';
  import { Ellipsis } from 'bkui-vue/lib/icon';
  import { useRoute } from 'vue-router';
  import { GetEnvApmOutput } from '~/@types/v1/bkintegrations-bkmonitor';
  import { EnvOutput } from '~/@types/v1/env';
  import { BkintegrationsBkmonitorService } from '~/api/modules/v1';
  import { type ApmQueryParams, DEFAULT_APM_CONFIG } from '~/common/const';
  import MonitorIframe from '~/components/monitor-iframe.vue';
  import SlideDetail from '~/components/slide-detail.vue';
  import { envTypeMap, envTypeTagClassMap } from '~/composables/use-env-manager';
  import { useMonitorIframe } from '~/composables/use-monitor-iframe';
  import { useUrlQuerySync } from '~/composables/use-url-query-sync';

  import ApmInstance from './apm-instance.vue';
  import BasicInfo from './basic-info.vue';
  import Setting from './setting.vue';

  // 定义所有可用的 Tab 值
  const ACTIVE_TABS = [
    'applicationList',
    'basicInfo',
    'canaryStrategy',
    'laneConfig',
    'laneSettings',
    'modules',
    'observability',
    'overview',
    'setting',
    'trafficManagement',
  ] as const;

  // 从常量数组推导类型
  type ActiveTabType = (typeof ACTIVE_TABS)[number];

  interface IProps {
    activeTab?: ActiveTabType;
    data: EnvOutput;
  }
  const props = defineProps<IProps>();
  const emits = defineEmits(['update', 'delete']);

  const route = useRoute();

  // monitor-iframe 公共能力：URL 构建（type=application，前缀含 apm_submenu，内部自取 bizId）
  const { buildIframeUrl } = useMonitorIframe('application', { apm_submenu: 0, needMenu: false });

  const slideDetailRef = ref<InstanceType<typeof SlideDetail>>();
  // iframe URL：初始化窗口期（observabilityQuery 就绪）构建一次；环境/APM 切换时重建（页面切换语义，整体 reload）
  const iframeUrl = ref('');

  // 侧滑状态与 URL query 双向同步锚定：activeTab 记忆 Tab；active 由列表页 env.vue 行点击统一写回（避免两处争写）
  const { fields } = useUrlQuerySync({
    activeTab: {
      queryKey: 'activeTab',
      data: {
        allowed: ACTIVE_TABS,
        default: props.activeTab || 'basicInfo',
      },
    },
    apmQuery: {
      queryKey: 'apmQuery',
      data: {
        default: '',
      },
    },
  });
  const activeTabName = fields.activeTab;

  const tabKey = computed(() => props.data?.createdAt?.toString() || Date.now().toString());

  const workspace = computed<string>(() => route.params.space as string);

  const envTypeConfig = computed(() => {
    if (props.data?.type && envTypeMap[props.data?.type]) {
      return envTypeMap[props.data.type];
    }
    return {} as { name: string; theme: string };
  });

  const currentApm = ref<GetEnvApmOutput | null>(null);
  const dropdownRef = ref<InstanceType<typeof Dropdown>>();

  // 获取当前环境关联的 APM 实例
  async function getEnvApm() {
    if (activeTabName.value !== 'observability') return;

    const envID = props.data.id;
    if (!envID) {
      currentApm.value = null;
      return;
    }

    currentApm.value = null;
    const apm = await BkintegrationsBkmonitorService.getEnvApm({ envID }, { interceptorErr: false }).catch(() => null);
    if (props.data.id === envID) {
      currentApm.value = apm;
    }
  }

  // 优先当前环境绑定的 APM name，没有则使用环境名
  const apmAppName = computed(() => {
    return currentApm.value?.name || props.data.name;
  });

  // 解析 URL 中的 iframe 观测参数（route-change 同步回写），异常/非对象/数组（如 JSON.parse('[1,2]')）返回空
  const parsedApmQuery = computed<ApmQueryParams>(() => {
    const raw = fields.apmQuery.value;
    if (!raw) return {};
    try {
      const parsed = JSON.parse(raw);
      return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? (parsed as ApmQueryParams) : {};
    } catch {
      return {};
    }
  });

  // 观测参数：其余字段（含 filter-service_name，iframe 内联动回写后经初始化 URL 恢复）由 URL 恢复，
  // filter-app_name 恒用当前环境派生值（避免 URL 残留旧环境的 app_name 覆盖）
  const observabilityQuery = computed(() => {
    const { 'filter-app_name': _queryAppName, ...restApmQuery } = parsedApmQuery.value;
    return {
      ...DEFAULT_APM_CONFIG,
      dashboardId: 'overview',
      ...restApmQuery,
      'filter-app_name': apmAppName.value ?? '',
    };
  });

  // 删除环境
  function handleDeleteEnv() {
    emits('delete', props.data);
    dropdownRef.value?.popoverRef?.hide();
  }

  /** iframe 内部状态变化：同步完整观测参数到 URL apmQuery 字段（供刷新/分享恢复） */
  function handleRouteChange(payload: { hash: string; href: string; query: Record<string, unknown> }) {
    const { query } = payload;
    if (!query || Object.keys(query).length === 0) return;
    fields.apmQuery.value = JSON.stringify(query);
  }

  function handleUpdate(row: EnvOutput) {
    emits('update', row);
  }

  // 显示面板
  function show() {
    slideDetailRef.value?.show();
  }

  watch([activeTabName, () => props.data?.id], getEnvApm, { immediate: true });

  // 派生参数（当前 APM 应用）变化：合并回写 apmQuery（跨环境不残留旧应用），并置空 iframeUrl 触发整体 reload
  // 仅当 appName 确实发生切换（prev 有值且不同）时才置空；首次由空变有效时 iframeUrl 本就为空，置空为 no-op，显式跳过语义更清晰
  watch(apmAppName, (appName, prevAppName) => {
    if (!appName) return;
    if (fields.apmQuery.value) {
      fields.apmQuery.value = JSON.stringify({ ...parsedApmQuery.value, 'filter-app_name': appName });
    }
    if (prevAppName && prevAppName !== appName) {
      iframeUrl.value = '';
    }
  });

  // 构建 iframe URL：filter-app_name 非空（currentApm 就绪）后构建一次；iframeUrl 被置空（环境/APM 切换）后再次构建
  // buildUrlSeq 序号：快速连续切换时丢弃过期构建结果，防止旧环境 URL 覆盖新状态
  let buildUrlSeq = 0;
  watch(
    () => observabilityQuery.value['filter-app_name'],
    appName => {
      if (iframeUrl.value || !appName) return;
      const seq = ++buildUrlSeq;
      buildIframeUrl(observabilityQuery.value).then(url => {
        if (seq === buildUrlSeq) {
          iframeUrl.value = url;
        }
      });
    },
    { immediate: true },
  );

  defineExpose({
    show,
  });
</script>
<style lang="postcss" scoped>
  :deep(.app-detail-tab) {
    font-size: 14px;
    height: calc(100% - 82px);
    .bk-tab-content {
      display: flex;
      flex-direction: column;
      flex-basis: calc(100% - 42px);
      overflow-y: auto;
    }
    .bk-tab-content > .bk-tab-panel {
      flex: 1;
      min-height: 0;
    }
  }
</style>
