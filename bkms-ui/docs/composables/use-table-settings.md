# `useTableSettings` 表格列设置 Hook

> 源码：`src/composables/use-table-settings.ts`

## 这个 Hook 做什么

表格右上角有"列设置"，用户可以自己勾选显示哪些列、调整行高。这个 Hook 就是管这件事的：

1. 用户调完的设置会**存下来**，下次打开页面自动恢复；
2. 告诉表格"当前该显示哪些列"（`settings`），并在用户改动时去保存（`handleSettingChange`）。

## 数据怎么存

**不存完整勾选，只存"和默认不一样的地方"**，避免存一堆重复数据。

比如默认勾选 A、B、C 三列，用户把 B 取消、勾上了 D，那就只记：

- `hidden: ['B']` —— 默认勾选的，但被用户取消的
- `shown: ['D']` —— 默认没勾的，但被用户勾上的
- `size` —— 用户选的行高

最后显示哪些列 = 默认勾选 − hidden + shown + 不可隐藏列（disabled），去重后得出。

存的位置是 `uiStore.preferences`，key 是 `table:${tableId}`（比如 `table:deploy-overview`），刷新页面不会丢。

## 几个设计细节

- **没改动就不留记录**：用户什么也没调，这条记录会被自动清掉，不存垃圾。
- **格式不对就按默认来**：如果本地存的数据结构不合法，直接当作没改过，显示默认列。
- **行高 `small` 不存**：`small` 就是表格默认行高，等于没个性化。否则用户只是点开一下设置面板、啥都没动，也会被存进去，导致"没改动就清除"失效。
- **不可隐藏列（disabled）**：像"环境"这种必须展示的列，传进 `disabled` 后用户就藏不掉了，也不会参与差异计算。

## 怎么用

```ts
const { settings, handleSettingChange } = useTableSettings(tableId, {
  defaultChecked: ['id', 'image', 'ip'], // 默认勾选哪些列
  disabled: ['id'],                       // 哪些列不许隐藏（可选）
});
```

模板里绑给表格：

```vue
<Table
  :settings="settings"
  :show-settings="true"
  @setting-change="handleSettingChange"
/>
```

两个参数：

| 参数 | 说明 |
|------|------|
| `tableId` | 给这张表起个唯一的名字（支持 `ref`/`computed`）。**同一个 id 就是同一份设置**；页面里 v-for 渲染多张表时，每张要传不同的 id，否则设置会互相串。 |
| `disabled` | 这些列强制显示，用户隐藏不了，也不参与差异计算。 |

## 现在谁在用

| 页面 | tableId | 默认勾选 | 不可隐藏 |
|------|---------|----------|----------|
| `deploy-overview.vue` | `'deploy-overview'`（固定） | 环境、类型、状态等 6 列 | 环境列 |
| `instance-table.vue` | `instance-table-${环境名}`（随环境变） | id、镜像、IP 等 9 列 | id 列 |

`instance-table` 页面每个环境渲染一张表格，所以 tableId 里拼了环境名——在环境 A 调列设置，不会影响环境 B。
