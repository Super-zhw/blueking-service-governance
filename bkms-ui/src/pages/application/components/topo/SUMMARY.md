# G6 拓扑图模块开发总结

> 基于 `@antv/g6` v5 的 K8s 资源拓扑图，展示应用部署环境下的资源依赖关系。

---

## 一、架构概览

```
index.vue                  ← 页面入口（数据获取、骨架屏、布局编排）
├── TopologySearch         ← 搜索组件（下拉搜索 + 命中翻页 + 定位）
├── TopologyStatistics     ← 左侧过滤面板（状态筛选 + 资源类型筛选）
├── ResourceTopology       ← 核心 G6 画布组件
│   ├── TopologyToolbar    ← 缩放/缩略图工具栏
│   ├── TopologyContextMenu← 右键菜单
│   ├── EdgeTooltip        ← 辅助边浮动提示
│   └── topology-node.vue  ← Vue 自定义节点（g6-extension-vue）
└── NodeDetailSidebar      ← 节点详情侧栏（概览/事件/日志/YAML）
```

### 核心依赖

| 包名               | 用途                           |
| ------------------ | ------------------------------ |
| `@antv/g6`         | 图引擎（布局、渲染、交互）     |
| `g6-extension-vue` | Vue 组件作为 G6 自定义节点     |
| `@antv/g`          | 底层渲染引擎（Group、Rect 等） |

---

## 二、G6 核心概念与 API

### 2.1 Graph 实例化

```typescript
const graph = new Graph({
  container: domElement,
  autoFit: 'center',
  padding: [60, 40, 40, 40],
  zoomRange: [0.25, 1.5],
  data: { nodes, edges },
  node: { type, style, size, ports },
  edge: { type, style },
  layout: DAGRE_LAYOUT_OPTIONS,
  behaviors: [...],
  plugins: [...],
  animation: false,
});
```

**关键参数**：

- `autoFit: 'center'`：首次渲染居中适配
- `zoomRange`：限制缩放范围，避免用户缩放到极小/极大
- `animation: false`：禁用全局动画（拓扑图数据量大时动画反而碍事）

### 2.2 自定义扩展注册

G6 v5 使用 `register(ExtensionCategory, name, class)` 注册自定义扩展：

```typescript
register(ExtensionCategory.NODE, 'resource-node', VueNode); // Vue 节点
register(ExtensionCategory.EDGE, 'primary-edge', PrimaryEdge); // 主边
register(ExtensionCategory.EDGE, 'auxiliary-edge', AuxiliaryEdge); // 辅助边
register(ExtensionCategory.BEHAVIOR, 'show-auxiliary-edges', ShowAuxiliaryEdgesOnHover);
```

> **⚠️ 避坑**：`register` 是全局注册，重复注册会报错。使用 `extensionsRegistered` 标志位确保只注册一次。

### 2.3 坐标系

G6 有三层坐标：

| 坐标系                   | 说明                         | 获取方式                                 |
| ------------------------ | ---------------------------- | ---------------------------------------- |
| **画布坐标（Canvas）**   | 数据模型坐标，布局计算的结果 | `graph.getElementPosition(id)`           |
| **视口坐标（Viewport）** | 相对于 graph 容器 DOM 的坐标 | `graph.getViewportByCanvas(canvasPoint)` |
| **客户端坐标（Client）** | 相对于浏览器视窗的坐标       | `event.client.x / y`                     |

> **⚠️ 避坑**：tooltip / 右键菜单等需要定位的组件，需从画布坐标转为视口坐标。使用 `getViewportByCanvas` 而非 `getCanvasByViewport`（后者是反向转换）。

### 2.4 元素状态（State）

G6 的 State 是一个字符串数组，如 `['selected']`、`['focused', 'highlighted']`。

```typescript
// 设置状态（覆盖式，注意会清空之前的状态！）
graph.setElementState(id, ['selected']);

// 增量添加状态（需手动合并）
const current = graph.getElementState(id);
graph.setElementState(id, [...current, 'active']);

// 增量移除
graph.setElementState(
  id,
  current.filter(s => s !== 'active'),
);

// 批量设置（性能更好）
await graph.setElementState(stateMap, false);
```

> **⚠️ 避坑**：`setElementState(id, states)` 是**全量覆盖**，不是增量合并。如果边已有 `['selected']`，直接 `setElementState(id, ['active'])` 会丢掉 `selected`。必须先 `getElementState` 再合并。

本项目中使用的状态：

- `selected`：点击选中节点
- `focused`：搜索当前命中项
- `highlighted`：搜索全部命中项
- `active`：辅助边 hover 高亮

### 2.5 数据更新策略

G6 提供多种更新方式，选择取决于场景：

| 方法                                            | 触发布局    | 适用场景                                   |
| ----------------------------------------------- | ----------- | ------------------------------------------ |
| `graph.render()`                                | ✅ 全量重排 | 数据大变（初始化/全量刷新）                |
| `graph.draw()`                                  | ❌          | 只重绘不重排（样式/可见性变化）            |
| `graph.layout()`                                | ✅          | 只重排不重绘                               |
| `addNodeData / updateNodeData / removeNodeData` | ❌          | 增量更新（需手动调 `draw()` / `layout()`） |

> **⚠️ 避坑**：增量更新后必须手动调 `await graph.draw()` 才能看到变化。如果有新增/删除节点还需 `await graph.layout()` 重排。

本项目使用 **diff 增量更新**策略（`diffArrayFast`），避免全量 `render()` 导致的布局抖动。

### 2.6 可见性控制

```typescript
graph.showElement(id, animate?);   // 显示
graph.hideElement(id, animate?);   // 隐藏（不删除数据）
graph.getElementVisibility(id);    // 获取可见状态
```

> **⚠️ 避坑**：`showOnlyNodeIds` 为空数组时表示"不过滤"（显示全部），但传给 G6 时空 `Set` 会导致所有节点被隐藏。本项目用 `null` 表示不过滤：

```typescript
const visibleNodeIds = computed(() => {
  const ids = props.showOnlyNodeIds;
  return ids.length > 0 ? new Set(ids) : null; // null = 不过滤
});
```

---

## 三、功能模块详解

### 3.1 自定义节点（VueNode + Vue 组件）

使用 `g6-extension-vue` 将 Vue 组件渲染为 G6 节点：

```typescript
node: {
  type: CUSTOM_NODE_TYPE,  // 'resource-node' → 注册为 VueNode
  style: {
    component: (data) => h(TopoNodeComponent, { data, onToggleCollapse }),
    size: [NODE_WIDTH, NODE_HEIGHT],  // 240 × 48
  },
}
```

**超采样解决文字模糊**：

G6 Canvas 中 HTML 节点通过 CSS `transform` 缩放，放大后文字会模糊。解决方案：

```css
/* topology-node.vue */
zoom: 4; /* 4 倍物理像素渲染 */
transform: scale(0.25); /* 缩回原始视觉大小 */
transform-origin: top left;
```

这样画布放大 4 倍以内，文字仍然清晰。常量 `NODE_SCALE_FACTOR = 4`。

### 3.2 自定义边（custom-edge.ts）

#### 3.2.1 路径（Path）基础概念

Canvas 路径是一系列绘图指令的集合，用于描述从起点到终点的图形轮廓。在 SVG 和 Canvas 2D 中，路径使用 **Path 命令** 来定义。

#### 3.2.2 常用路径命令

| 命令 | 含义             | 参数说明                                             |
| ---- | ---------------- | ---------------------------------------------------- |
| `M`  | Move To          | 移动到起点 `(x, y)`，不绘制线条                      |
| `L`  | Line To          | 从当前点绘制直线到 `(x, y)`                          |
| `A`  | Arc              | 绘制圆弧 `(rx, ry, rotation, largeArc, sweep, x, y)` |
| `Q`  | Quadratic Bezier | 二次贝塞尔曲线 `(cx, cy, x, y)`                      |
| `C`  | Cubic Bezier     | 三次贝塞尔曲线 `(c1x, c1y, c2x, c2y, x, y)`          |
| `Z`  | Close Path       | 闭合路径                                             |

#### 3.2.3 Arc 命令详解（重点）

`A` 命令是绘制圆弧的核心，包含 7 个参数：

```typescript
['A', rx, ry, rotation, largeArc, sweep, x, y];
```

- **rx, ry**: X/Y 轴半径（椭圆时不同，正圆时相同）
- **rotation**: 椭圆旋转角度（度）
- **largeArc**: 是否选择大弧（0=小弧，1=大弧）
- **sweep**: 弧线方向（0=逆时针，1=顺时针）
- **x, y**: 圆弧终点坐标

---

## G6 中如何自定义路径

#### 3.2.4 G6 边（Edge）的继承体系

在 AntV G6 中，自定义边通常继承以下基类：

```typescript
// 水平三次贝塞尔曲线（适用于水平布局的树）
class AuxiliaryEdge extends CubicHorizontal {}

// 折线（适用于直角连线场景）
class PrimaryEdge extends Polyline {}
```

#### 3.2.5 核心方法：`getKeyPath()`

`getKeyPath()` 方法返回 `PathArray`，定义边的几何形状：

```typescript
protected getKeyPath(): PathArray {
  const { sourceNode, targetNode } = this;
  const { width, height } = sourceNode.getBBox();
  const [x1, y1] = sourceNode.getPosition();
  const [x2, y2] = targetNode.getPosition();
  // ... 计算路径点
}
```

#### 3.2.6 路径点计算详解

以微服务拓扑中的主边为例，路径结构如下：

```
源节点右侧 ──→ 水平线段 ──→ 圆弧转角 ──→ 竖直主干 ──→ 圆弧转角 ──→ 水平线段 ──→ 目标节点左侧
```

**关键坐标点计算：**

```typescript
// 起点：源节点右侧中心
const startX = x1 + width;
const startY = y1 + height / 2 - offset;

// 终点：目标节点左侧中心
const endX = x2;
const endY = y2 + height / 2 - offset;

// 竖直主干的 X 坐标（源和目标中间）
const midX = (x1 + x2) / 2 + width / 2 - lineWidth / 2;
```

#### 3.2.7 完整路径构建

```typescript
const path: PathArray = [
  // 1. 起点：移动到源节点右侧
  ['M', startX, startY],

  // 2. 水平线到第一个转角前（预留圆弧空间）
  ['L', midX - effectiveR, startY],

  // 3. 第一个圆弧转角（水平 → 竖直）
  //    rx=ry=r, rotation=0, largeArc=0, sweep=dirY>0?1:0
  ['A', effectiveR, effectiveR, 0, 0, dirY > 0 ? 1 : 0, midX, startY + dirY * effectiveR],

  // 4. 竖直线到第二个转角前
  ['L', midX, endY - dirY * effectiveR],

  // 5. 第二个圆弧转角（竖直 → 水平）
  ['A', effectiveR, effectiveR, 0, 0, dirY > 0 ? 0 : 1, midX + effectiveR, endY],

  // 6. 水平线到终点
  ['L', endX, endY],
];
```

#### 3.2.8 圆弧方向控制（sweep 参数）

| 方向     | sweep 值 | 说明       |
| -------- | -------- | ---------- |
| 向下转弯 | `1`      | 顺时针方向 |
| 向上转弯 | `0`      | 逆时针方向 |

```typescript
const dirY = dy > 0 ? 1 : -1; // 判断目标在上方还是下方

// 第一个转角：根据方向选择 sweep
['A', r, r, 0, 0, dirY > 0 ? 1 : 0, midX, startY + dirY * r][
  // 第二个转角：反向 sweep
  ('A', r, r, 0, 0, dirY > 0 ? 0 : 1, midX + r, endY)
];
```

#### 3.2.9 样式与渲染

```typescript
protected getKeyStyle(attributes: Required<PolylineStyleProps>) {
  return {
    ...super.getKeyStyle(attributes),
    lineWidth: PrimaryEdge.lineWidth,
    stroke: '#ABB5CC',
  };
}

render(attributes: Required<PolylineStyleProps>, container: Group) {
  super.render(attributes, container);
  this.drawTargetArrow({
    ...attributes,
    endArrow: true,  // 绘制箭头
  });
}
```

---

## 微服务拓扑中的 Tree 路径实践

#### 3.2.10 业务场景

在微服务资源拓扑中，需要表达两种关系：

- **主关系（Primary）**：服务调用链的主路径，使用直角折线
- **辅助关系（Auxiliary）**：非核心依赖，使用虚线贝塞尔曲线

#### 3.2.11 辅助边实现（CubicHorizontal）

```typescript
class AuxiliaryEdge extends CubicHorizontal {
  protected getKeyStyle(attributes: Required<CubicHorizontalStyleProps>) {
    return {
      ...super.getKeyStyle(attributes),
      stroke: '#ABB5CC',
      lineWidth: 2,
      lineDash: [3, 3], // 虚线效果
    };
  }

  render(attributes: Required<CubicHorizontalStyleProps>, container: Group) {
    super.render(attributes, container);
    this.drawTargetArrow({ ...attributes, endArrow: true });
  }
}
```

#### 3.2.12 动态边管理（Behavior）

辅助边不参与布局，仅在交互时动态添加/移除：

```typescript
private async addAuxiliaryEdges(nodeId: string) {
  const { graph } = this.context;
  const allAuxEdges = this.options.auxiliaryEdges ?? [];

  // 过滤：只保留与当前节点相关且两端可见的边
  const related = allAuxEdges.filter(e => {
    if (e.source !== nodeId && e.target !== nodeId) return false;
    const sourceVisible = graph.getElementVisibility(e.source as string) !== 'hidden';
    const targetVisible = graph.getElementVisibility(e.target as string) !== 'hidden';
    return sourceVisible && targetVisible;
  });

  this.activeEdgeIds = related.map(e => e.id!);
  graph.addEdgeData(related);  // 动态添加
  await graph.draw();
}
```

#### 3.2.13 注册扩展

```typescript
export function registerCustomExtensions() {
  register(ExtensionCategory.EDGE, 'primary-edge', PrimaryEdge);
  register(ExtensionCategory.EDGE, 'auxiliary-edge', AuxiliaryEdge);
  register(ExtensionCategory.BEHAVIOR, 'show-auxiliary-edges', ShowAuxiliaryEdgesOnHover);
  register(ExtensionCategory.BEHAVIOR, 'highlight-auxiliary-edge', HighlightAuxiliaryEdgeOnHover);
}
```

#### 3.2.14 使用示例

```typescript
const graph = new Graph({
  edge: {
    type: 'primary-edge', // 使用自定义主边
  },
  behaviors: [
    'show-auxiliary-edges', // hover 显示辅助边
    'highlight-auxiliary-edge', // 高亮交互
  ],
});
```

---

## 总结

| 要点         | 说明                                        |
| ------------ | ------------------------------------------- |
| **路径命令** | M/L/A 是基础，A 命令控制圆弧方向            |
| **继承选择** | CubicHorizontal 用于曲线，Polyline 用于折线 |
| **动态管理** | 通过 Behavior 实现按需渲染，提升性能        |
| **坐标计算** | 基于节点 BBox 和 Position 计算连接点        |

这篇文章涵盖了从 Canvas 基础到 G6 实践，再到微服务拓扑场景的完整路径自定义方案。

#### PrimaryEdge（主边）

继承 `Polyline`，手动绘制带圆弧转角的折线路径：

```
源节点右侧 → 水平 → 圆弧转角 → 竖直 → 圆弧转角 → 水平 → 目标节点左侧
```

关键实现：

- `getKeyPath()`：手动计算 SVG PathArray，包含圆弧（`A` 命令）
- `effectiveR`：当垂直距离不足两倍圆弧半径时，动态缩小半径
- `drawTargetArrow()`：渲染目标端箭头

#### AuxiliaryEdge（辅助边）

继承 `CubicHorizontal`（水平三次贝塞尔曲线），虚线样式：

```typescript
protected getKeyStyle(attributes) {
  const isActive = (attributes as any).states?.includes('active');
  return {
    stroke: isActive ? '#7E8EAD' : '#ABB5CC',
    lineWidth: 2,
    lineDash: isActive ? undefined : [3, 3],  // hover 时变实线
  };
}
```

> **⚠️ 避坑**：自定义边中读取 state 需要通过 `(attributes as any).states`，G6 v5 的类型定义未包含此字段。

### 3.3 自定义 Behavior（ShowAuxiliaryEdgesOnHover）

辅助边不参与布局，仅在 hover/click 节点时**按需动态添加**到画布：

```
hover 节点 → addAuxiliaryEdges() → graph.addEdgeData() → graph.draw()
leave 节点 → removeActiveAuxiliaryEdges() → graph.removeEdgeData() → graph.draw()
click 节点 → 切换 selectedNodeId（锁定/解锁辅助边）
click 画布 → 清除 selectedNodeId + 移除辅助边
```

**关键设计**：

- `selectedNodeId` 是模块级变量（非响应式），用于在 click 和 hover 之间协调
- hover 折叠按钮（`.custom-collapse`）时不触发辅助边：通过 `domTarget?.classList?.contains('custom-collapse')` 判断
- 移除前先检查边是否确实存在于画布：`existing.has(id)`

> **⚠️ 避坑**：辅助边是动态添加的，不能纳入 `updateGraphEdges()` 的 diff 逻辑，否则会被误判为"removed"而删掉。当前 `updateGraphEdges` 只处理主边。

### 3.4 Minimap（缩略图）

```typescript
plugins: [{
  type: 'minimap',
  size: [300, 180],
  containerStyle: { background: '#242A35' },
  filter: (id, elementType) => ...,  // 只显示可见节点
  shape: (id, elementType, element) => {  // 自定义缩略图节点形状
    // 返回 Group：白色底板 + 左侧状态色块
  },
}]
```

**缩略图节点样式**：使用 `@antv/g` 的 `Group` + `Rect` 组合：

```typescript
const group = new Group();
// 白色圆角底板
group.appendChild(new Rect({ style: { width: NODE_WIDTH, height: NODE_HEIGHT, fill: '#FFFFFF', radius: 10 } }));
// 内嵌左侧状态色块
group.appendChild(new Rect({ style: { x: 6, y: 6, width: 36, height: 36, fill: statusColor, radius: 6 } }));
```

**minimap 可视区域的"开窗"效果**（CSS hack）：

```css
:deep(.g6-minimap) {
  div {
    background: transparent !important;
    box-shadow: 0 0 10000px 10000px #0000002d; /* 超大阴影模拟暗色遮罩 */
  }
}
```

### 3.5 辅助边 Tooltip

**放弃 G6 内置 tooltip 插件**（跟随鼠标，无法固定在边中心），改用自定义 Vue 组件：

1. `edge:pointerenter` → 计算边渲染包围盒中心 → `getViewportByCanvas` 转视口坐标
2. `EdgeTooltip` 组件接收 `x/y/visible/relation/reason`，通过 CSS `transform: translate(calc(-100% - 12px), -50%)` 定位在边中心左侧
3. 右侧 CSS 三角形箭头指向边

> **⚠️ 避坑**：不要用源/目标节点位置的中点作为边中心，辅助边是曲线（CubicHorizontal），视觉中点和端点中点不同。用 `graph.getElementRenderBounds(edgeId)` 获取边的实际渲染包围盒中心。

### 3.6 搜索与定位

搜索流程：

```
输入 → matchedNodes 过滤 → 下拉面板展示
Enter/点击搜索 → emit('update:selectedIds', ids) + emit('locate', nodeId)
点击下拉项 → 回填名称 + 定位
上一个/下一个 → emit('locate', nodeId)
```

G6 侧的状态同步：

```typescript
watch([parsedSelectedNodeIds, focusedNodeId], () => {
  // 增量更新，保留非搜索相关状态
  const next = graph.getElementState(id).filter(s => s !== 'focused' && s !== 'highlighted');
  if (id === focusedId) next.push('focused');
  else if (highlightedSet.has(id)) next.push('highlighted');
  stateMap[id] = next;
  await graph.setElementState(stateMap, false); // 批量设置
});
```

### 3.7 左侧过滤面板

状态筛选 + 资源类型筛选，联动控制可见节点：

| 场景              | visibleNodeIds                            |
| ----------------- | ----------------------------------------- |
| 全选 + 全部状态   | `[]`（不过滤）                            |
| 全选 + 非全部状态 | 该状态所有节点 ID                         |
| 部分选中          | 匹配节点 ID                               |
| 全部取消          | `['__none__']`（占位 ID，让所有节点隐藏） |

> **⚠️ 避坑**：`visibleNodeIds` 为空数组在 G6 侧被解读为"不过滤"，需要用占位 ID `['__none__']` 来表示"全部隐藏"。

---

## 四、疑难点与避坑指南

### 4.1 HTML 节点文字模糊

**问题**：G6 Canvas 渲染器中，HTML 节点通过 `transform: scale()` 缩放，放大后文字模糊。

**方案**：CSS `zoom` + `transform: scale(1/zoom)` 超采样。

```typescript
export const NODE_SCALE_FACTOR = 4; // 4 倍超采样
```

### 4.2 `setElementState` 覆盖而非合并

**问题**：`graph.setElementState(id, ['active'])` 会清除之前的 `selected` 状态。

**方案**：增量操作时先 `getElementState` 再合并/过滤。

### 4.3 辅助边被 diff 逻辑误删

**问题**：`updateGraphEdges` 做 diff 时，hover 动态添加的辅助边不在 `graphEdges` 中，会被判定为 "removed" 而删除。

**方案**：`updateGraphEdges` 只处理主边（`isPrimary`），辅助边由 Behavior 自行管理。

### 4.4 空数组 = 不过滤 vs 全部隐藏

**问题**：`showOnlyNodeIds: []` 传到 G6 后，`visibleNodeIds` 计算为 `null`（不过滤），但业务上"全部取消选中"应表示"隐藏所有"。

**方案**：全部取消时传入占位 ID `['__none__']`，确保 G6 侧认为"有过滤条件但无节点匹配"。

### 4.5 Minimap 节点形状默认为单色矩形

**问题**：默认 minimap 节点只显示一个带颜色的矩形，无法区分"左侧色块 + 右侧内容"。

**方案**：通过 `shape` 回调返回 `Group`（白色底板 + 内嵌色块）。

### 4.6 Tooltip 跟随鼠标而非固定在边中心

**问题**：G6 内置 tooltip 插件跟随鼠标位置，无法固定在边中心。

**方案**：移除内置插件，改用 Vue 组件 + `edge:pointerenter` 事件 + `getElementRenderBounds` + `getViewportByCanvas` 手动定位。

### 4.7 自定义边中读取 State

**问题**：G6 v5 的 `CubicHorizontalStyleProps` 类型不包含 `states` 字段。

**方案**：`(attributes as any).states?.includes('active')`，需类型断言。

### 4.8 折叠按钮触发 hover

**问题**：hover 节点上的折叠按钮时，会意外触发辅助边显示。

**方案**：在 `onNodePointerEnter` 中检测 `domTarget?.classList?.contains('custom-collapse')`，若是则跳过。

### 4.9 `render()` 导致布局抖动

**问题**：频繁调用 `graph.render()` 会导致全量布局重算，视觉上产生抖动。

**方案**：优先使用增量更新（`addNodeData / updateNodeData / removeNodeData` + `draw()`），只在结构变化时才调 `layout()`。

### 4.10 坐标转换方向搞反

**问题**：`getCanvasByViewport` 和 `getViewportByCanvas` 名字容易混淆。

**口诀**：

- `getViewportByCanvas(canvas)` = 画布坐标 → 视口坐标（**用于 tooltip 定位**）
- `getCanvasByCanvas(viewport)` = 视口坐标 → 画布坐标（反向）

---

## 五、性能优化

1. **`shallowRef` 存储 Graph 实例**：避免 Vue 深度代理 G6 内部对象
2. **`useWorker: true`**：布局计算使用 Web Worker，避免阻塞主线程
3. **diff 增量更新**：只 add/update/remove 变化的数据，而非全量 render
4. **`animation: false`**：禁用全局动画，减少不必要的帧计算
5. **4 秒轮询 + shallowRef**：`useIntervalFn(handleGetResourceTopology, 4000)` + `shallowRef` 避免每次轮询触发不必要的响应式更新
6. **辅助边按需加载**：不参与布局，hover 时才添加，减少渲染负担

---

## 六、文件职责速查

| 文件                        | 职责                                                              |
| --------------------------- | ----------------------------------------------------------------- |
| `index.vue`                 | 页面入口：数据获取、骨架屏、布局编排                              |
| `resource-topology.vue`     | G6 画布核心：Graph 实例化、事件、数据同步、可见性                 |
| `custom-edge.ts`            | 自定义边（PrimaryEdge / AuxiliaryEdge）+ 自定义 Behavior          |
| `constants.ts`              | 常量（节点尺寸、状态颜色、Kind 图标映射、布局配置、资源分类）     |
| `types.ts`                  | 类型定义（TopoNodeData、CategoryGroup、KindGroup、NodeStatus 等） |
| `topology-node.vue`         | Vue 自定义节点组件（超采样、状态样式、折叠按钮）                  |
| `edge-tooltip.vue`          | 辅助边浮动提示（定位 + 箭头 + 动画）                              |
| `topology-toolbar.vue`      | 缩放/缩略图工具栏                                                 |
| `topology-search.vue`       | 搜索组件（下拉搜索 + 命中翻页 + 定位）                            |
| `topology-statistics.vue`   | 左侧过滤面板（状态筛选 + 资源类型筛选）                           |
| `topology-context-menu.vue` | 右键上下文菜单                                                    |
| `topology-status-icon.vue`  | 状态图标组件                                                      |
| `node-detail-sidebar.vue`   | 节点详情侧栏                                                      |
| `detail-overview.vue`       | 概览 Tab                                                          |
| `detail-events.vue`         | 事件 Tab                                                          |
| `detail-log.vue`            | 日志 Tab                                                          |
| `detail-yaml.vue`           | YAML Tab                                                          |
| `resource-list.vue`         | 资源列表模式                                                      |
