# 主题与前端

XLPanel 采用「配置驱动主题」的思路，后续可在前端层加载不同主题资源。

## 推荐策略

1. 前端构建为独立包，例如 `frontend/themes/<theme-name>`。
2. 通过 `core.DefaultConfig.DefaultTheme` 设置默认主题。
3. 对外 API 暴露主题信息，前端可根据租户/客户偏好切换。

## 可替换主题示例

- classic
- modern
- dark

## 后续可扩展

- 主题配置中心（数据库表 + 管理接口）
- 多租户主题覆盖策略
