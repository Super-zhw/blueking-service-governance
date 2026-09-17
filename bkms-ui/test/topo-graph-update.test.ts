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
/**
 * 拓扑图增量更新契约测试（jsdom + G6 替身）。
 *
 * 背景（--story=138007525）：切换环境时拓扑图「崩坏」——节点重合、连线乱飞。
 * 根因：nodes / edges 两个 watcher 各自 diff + draw + layout。节点 watcher 先跑，
 * removeNodeData 会**级联删掉关联边**（见 @antv/g6 src/runtime/data.ts removeNodeData），
 * 于是图进入「有节点、无新边」的中间态；而布局是 indented（缩进树），完全依赖边
 * 建立父子层次 → 所有节点被当成 N 个独立的根 → 摆到同一列/原点 → 视觉上就是「崩坏」。
 *
 * jsdom 无布局引擎，测不了坐标。本套用例在**调用契约层**提供毫秒级回归，断言：
 *   1. 任何一次 draw()/layout() 之前，节点与边的增删必须已全部落图（禁止半图提交）；
 *   2. 一次数据变更只允许一次 draw + 一次 layout（禁止双份布局并发）；
 *   3. 空图（切换环境的 reset 中间态）不执行布局；
 *   4. 数据无实质变化时（30s 轮询）不触发任何绘制。
 * 只要有人把「先节点后边 → 一次 draw → 一次 layout」拆回双 watch，这里立刻变红。
 */
import { nextTick } from 'vue';

import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ResourceTopology from '~/pages/application/components/topo/resource-topology.vue';

import type { TopologyEdge, TopologyNode } from '~/@types/topology';

// ── G6 替身：记录调用序列 + 每次布局时给图状态拍快照 ──
// 必须在 vi.hoisted 内定义：vi.mock 的 factory 会在静态 import 之前被提升执行。
const g6 = vi.hoisted(() => {
  type NodeLike = { children?: string[]; data?: Record<string, unknown>; id: string };
  type EdgeLike = { data?: Record<string, unknown>; id: string; source: string; target: string };
  type Snapshot = { components: number; edgeCount: number; nodeCount: number };

  /** 无向连通分量数：indented 布局下边缺失会让每个节点各自成根 → 分量数 = 节点数 */
  function countComponents(nodeIds: string[], edges: EdgeLike[]): number {
    const parent = new Map(nodeIds.map(id => [id, id]));
    const find = (a: string): string => {
      let root = a;
      while (parent.get(root) !== root) root = parent.get(root)!;
      let cursor = a;
      while (parent.get(cursor) !== root) {
        const next = parent.get(cursor)!;
        parent.set(cursor, root);
        cursor = next;
      }
      return root;
    };
    for (const edge of edges) {
      if (!parent.has(edge.source) || !parent.has(edge.target)) continue;
      const rootA = find(edge.source);
      const rootB = find(edge.target);
      if (rootA !== rootB) parent.set(rootA, rootB);
    }
    return new Set([...parent.keys()].map(find)).size;
  }

  const graphs: FakeGraph[] = [];

  class FakeGraph {
    private readonly options: Record<string, unknown>;
    private merge(target: Array<EdgeLike | NodeLike>, patch: Array<EdgeLike | NodeLike>) {
      for (const item of patch) {
        const index = target.findIndex(old => old.id === item.id);
        if (index >= 0) target[index] = { ...target[index], ...item };
      }
    }
    private snapshot(): Snapshot {
      return {
        components: countComponents(
          this.nodes.map(node => node.id),
          this.edges,
        ),
        edgeCount: this.edges.length,
        nodeCount: this.nodes.length,
      };
    }
    /** 调用序列（draw / layout / addXxxData / removeXxxData），用于断言「先补齐数据再提交绘制」 */
    calls: string[] = [];

    edges: EdgeLike[] = [];

    nodes: NodeLike[] = [];

    /** 每次布局（含 render 时的首次布局）的图状态快照 */
    readonly snapshots: Snapshot[] = [];
    constructor(options: Record<string, unknown>) {
      this.options = options;
      const data = (options.data ?? {}) as { edges?: EdgeLike[]; nodes?: NodeLike[] };
      this.nodes = (data.nodes ?? []).map(node => ({ ...node }));
      this.edges = (data.edges ?? []).map(edge => ({ ...edge }));
      graphs.push(this);
    }
    // ── 数据增删改 ──
    addEdgeData(list: EdgeLike[]) {
      this.calls.push('addEdgeData');
      for (const item of list) this.edges.push({ ...item });
    }
    addNodeData(list: NodeLike[]) {
      this.calls.push('addNodeData');
      for (const item of list) this.nodes.push({ ...item });
    }
    collapseElement() {
      return Promise.resolve();
    }
    destroy() {
      this.calls.push('destroy');
    }
    async draw() {
      this.calls.push('draw');
    }
    expandElement() {
      return Promise.resolve();
    }
    focusElement() {
      return Promise.resolve();
    }
    getData() {
      return { edges: this.edges, nodes: this.nodes };
    }
    getEdgeData(id?: string) {
      return id ? this.edges.find(edge => edge.id === id) : this.edges;
    }
    getElementState() {
      return [] as string[];
    }
    getElementVisibility() {
      return 'visible';
    }
    getNodeData(id?: string) {
      return id ? this.nodes.find(node => node.id === id) : this.nodes;
    }
    getOptions() {
      return this.options as { behaviors?: unknown[] };
    }
    hideElement() {
      this.calls.push('hideElement');
    }
    async layout() {
      this.calls.push('layout');
      this.snapshots.push(this.snapshot());
    }
    off() {}
    on() {}
    removeEdgeData(ids: string[]) {
      this.calls.push('removeEdgeData');
      const set = new Set(ids);
      this.edges = this.edges.filter(edge => !set.has(edge.id));
    }
    /** 对齐 G6：删节点会级联删掉其关联边（data.ts removeNodeData → removeEdgeData(getRelatedEdgesData)） */
    removeNodeData(ids: string[]) {
      this.calls.push('removeNodeData');
      const set = new Set(ids);
      this.nodes = this.nodes.filter(node => !set.has(node.id));
      this.edges = this.edges.filter(edge => !set.has(edge.source) && !set.has(edge.target));
    }
    /** render 内部同样会执行一次布局，一并纳入契约 */
    render() {
      this.calls.push('render');
      this.snapshots.push(this.snapshot());
    }
    resize() {}
    setBehaviors() {}
    setElementState() {
      return Promise.resolve();
    }
    showElement() {
      this.calls.push('showElement');
    }
    updateBehavior() {}

    updateEdgeData(list: EdgeLike[]) {
      this.calls.push('updateEdgeData');
      this.merge(this.edges, list);
    }

    updateNodeData(list: NodeLike[]) {
      this.calls.push('updateNodeData');
      this.merge(this.nodes, list);
    }
  }

  /** 自定义边继承 Polyline：只需可 new，测试路径不触达其绘制 API */
  class Polyline {}
  /** 自定义行为继承 BaseBehavior：同上 */
  class BaseBehavior {
    context: unknown;
    options: Record<string, unknown>;
    constructor(context: unknown, options: Record<string, unknown> = {}) {
      this.context = context;
      this.options = options;
    }
    destroy() {}
  }

  return {
    BaseBehavior,
    ExtensionCategory: { BEHAVIOR: 'behavior', EDGE: 'edge', NODE: 'node' },
    Graph: FakeGraph,
    Polyline,
    graphs,
    register: vi.fn(),
  };
});

vi.mock('@antv/g6', () => ({
  BaseBehavior: g6.BaseBehavior,
  ExtensionCategory: g6.ExtensionCategory,
  Graph: g6.Graph,
  Polyline: g6.Polyline,
  register: g6.register,
}));

// @antv/g 与 g6-extension-vue 仅在类型或回调中使用，测试路径不触达其 API
vi.mock('@antv/g', () => ({
  DisplayObject: class {},
  Group: class {
    appendChild() {}
  },
  Rect: class {},
}));
vi.mock('g6-extension-vue', () => ({ VueNode: {} }));

// 部分覆盖：保留真实 createI18n（组件链顶层可能创建实例），仅替换组件内 useI18n
vi.mock('vue-i18n', async importOriginal => ({
  ...(await importOriginal<object>()),
  useI18n: () => ({ t: (key: string) => key }),
}));

// ── 测试数据：一棵单根树 App → Service → Deployment → Pod ×N ──
function buildTopology(prefix: string, podCount: number) {
  const nodes: TopologyNode[] = [
    { id: 'app-1', kind: 'App', name: 'app-1', status: 'Running' },
    { id: `svc-${prefix}`, kind: 'Service', name: `svc-${prefix}`, status: 'Running' },
    { id: `deploy-${prefix}`, kind: 'Deployment', name: `deploy-${prefix}`, status: 'Running' },
    ...Array.from({ length: podCount }, (_, index) => ({
      id: `pod-${prefix}-${index + 1}`,
      kind: 'Pod',
      name: `pod-${prefix}-${index + 1}`,
      status: 'Running',
    })),
  ];

  const edges: TopologyEdge[] = [
    { id: `e-app-svc-${prefix}`, isPrimary: true, sourceID: 'app-1', targetID: `svc-${prefix}` },
    {
      id: `e-svc-deploy-${prefix}`,
      isPrimary: true,
      sourceID: `svc-${prefix}`,
      targetID: `deploy-${prefix}`,
    },
    ...Array.from({ length: podCount }, (_, index) => ({
      id: `e-deploy-pod-${prefix}-${index + 1}`,
      isPrimary: true,
      sourceID: `deploy-${prefix}`,
      targetID: `pod-${prefix}-${index + 1}`,
    })),
  ];

  return { edges, nodes };
}

/** 推平 Vue flush + 微任务链（updateGraph 是 async，watch 默认 pre flush） */
async function flush(rounds = 8) {
  for (let index = 0; index < rounds; index += 1) {
    await nextTick();
    await Promise.resolve();
  }
  await new Promise(resolve => setTimeout(resolve, 0));
}

const mounted: Array<ReturnType<typeof mount>> = [];

/** 断言：每一次布局时图数据都是自洽的（非空、有边、单根连通） */
function expectLayoutsConsistent(snapshots: Array<{ components: number; edgeCount: number; nodeCount: number }>) {
  for (const [index, snapshot] of snapshots.entries()) {
    expect(snapshot.nodeCount, `第 ${index + 1} 次布局时图为空，不应在空图上执行布局`).toBeGreaterThan(0);
    expect(snapshot.edgeCount, `第 ${index + 1} 次布局时边缺失，节点会全部退化成独立的根`).toBeGreaterThan(0);
    expect(snapshot.components, `第 ${index + 1} 次布局时图不连通（节点被拆成多棵树 → 重叠）`).toBe(1);
  }
}

async function mountTopology() {
  const envA = buildTopology('a', 2);
  const wrapper = mount(ResourceTopology, {
    attachTo: document.body,
    props: { edges: envA.edges, nodes: envA.nodes },
    shallow: true,
  });
  mounted.push(wrapper);
  await flush();
  return wrapper;
}

describe('拓扑图增量更新：draw/layout 时点的数据自洽性', () => {
  beforeEach(() => {
    g6.graphs.length = 0;
    g6.register.mockClear();
  });

  afterEach(() => {
    for (const wrapper of mounted.splice(0)) wrapper.unmount();
    document.body.innerHTML = '';
  });

  it('初始化：首次渲染时节点与边同时就位（单根连通）', async () => {
    await mountTopology();

    const graph = g6.graphs.at(-1)!;
    expect(graph.snapshots, '初始化只应产生一次布局').toHaveLength(1);
    expect(graph.snapshots[0]).toMatchObject({ components: 1, edgeCount: 4, nodeCount: 5 });
  });

  it('整体替换数据（切换环境）：数据补齐后才提交绘制，一次变更只 draw/layout 一次', async () => {
    const wrapper = await mountTopology();
    const graph = g6.graphs.at(-1)!;
    const before = graph.snapshots.length;

    graph.calls.length = 0;
    const envB = buildTopology('b', 3);
    await wrapper.setProps({ edges: envB.edges, nodes: envB.nodes });
    await flush();

    // 确实发生了重新布局
    expect(graph.snapshots.length).toBeGreaterThan(before);

    // 核心契约 1：所有增删必须先于任何一次 draw/layout 落图
    // 双 watch 实现下，节点 watcher 会在边补齐之前先 draw → 此处变红
    const firstCommit = Math.min(
      ...['draw', 'layout'].map(name => graph.calls.indexOf(name)).filter(index => index >= 0),
    );
    const lastMutation = Math.max(
      ...['addNodeData', 'addEdgeData', 'removeNodeData', 'removeEdgeData'].map(name => graph.calls.lastIndexOf(name)),
    );
    expect(firstCommit, 'draw/layout 必须发生在所有节点/边增删之后（禁止在半图状态下提交）').toBeGreaterThan(
      lastMutation,
    );

    // 核心契约 2：一次数据变更只允许一次 draw + 一次 layout
    expect(
      graph.calls.filter(call => call === 'draw'),
      '一次变更只应提交一次绘制',
    ).toHaveLength(1);
    expect(
      graph.calls.filter(call => call === 'layout'),
      '一次变更只应执行一次布局',
    ).toHaveLength(1);

    // 核心契约 3：每次布局时图数据自洽
    expectLayoutsConsistent(graph.snapshots);

    // 最终数据与新环境一致
    expect(graph.nodes.map(node => node.id).sort()).toEqual(
      ['app-1', 'deploy-b', 'pod-b-1', 'pod-b-2', 'pod-b-3', 'svc-b'].sort(),
    );
  });

  it('清空数据（切换环境的 reset 中间态）：空图不执行布局', async () => {
    const wrapper = await mountTopology();
    const graph = g6.graphs.at(-1)!;
    const before = graph.snapshots.length;

    graph.calls.length = 0;
    await wrapper.setProps({ edges: [], nodes: [] });
    await flush();

    expect(graph.snapshots.length, '清空后不应再产生布局（空图布局是纯粹的浪费且可能污染图状态）').toBe(before);
    expect(graph.calls.filter(call => call === 'layout')).toHaveLength(0);
    expectLayoutsConsistent(graph.snapshots);
  });

  it('数据无实质变化（30s 轮询）：不触发任何绘制', async () => {
    const wrapper = await mountTopology();
    const graph = g6.graphs.at(-1)!;

    graph.calls.length = 0;
    // 内容完全一致，仅对象引用变化（轮询每次返回新对象）
    const envA = buildTopology('a', 2);
    await wrapper.setProps({ edges: envA.edges, nodes: envA.nodes });
    await flush();

    expect(
      graph.calls.filter(call => call === 'draw'),
      '数据无变化时不应触发绘制',
    ).toHaveLength(0);
    expect(
      graph.calls.filter(call => call === 'layout'),
      '数据无变化时不应触发布局',
    ).toHaveLength(0);
  });
});
