# clash-configs-web

`clash-configs` 后端（Kotlin + Spring Boot）的前端管理界面，基于 **Vite + Vue 3 + TypeScript + shadcn-vue + Tailwind CSS**。

## 功能

- 登录（Spring Security 表单登录 + session cookie）
- 仪表盘：订阅流量用量汇总（上传 / 下载 / 总量 / 到期时间）
- 订阅配置管理：增删改查、手动更新、启用开关、更新周期（每天 / 每周）
- 合并配置管理：创建、编辑、刷新 Token、复制订阅链接
- 修改密码

## 目录结构

```
src/
├── api/            # axios 封装与后端接口
├── components/     # 布局与 shadcn-vue UI 组件
├── composables/    # useAuth 认证状态
├── lib/            # 工具函数（format、config、cn）
├── router/         # 路由与登录守卫
├── types/          # 类型定义
└── views/          # 页面
```

## 开发

```bash
npm install
npm run dev
```

开发服务器运行在 `http://localhost:5173`，通过 Vite 代理把 `/api/*` 转发到后端 `http://localhost:8080`（去掉 `/api` 前缀）。

后端默认账号：`admin` / `password`。

## 生产部署

前端构建产物为纯静态文件，部署时通过 Nginx 同源反代到后端，避免跨域：

```nginx
server {
    listen 80;
    server_name example.com;

    # 前端静态资源
    root /var/www/clash-configs-web/dist;
    index index.html;
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 后端 API（去掉 /api 前缀）
    location /api/ {
        proxy_pass http://127.0.0.1:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 公开订阅接口（Clash 客户端直接访问）
    location /configs {
        proxy_pass http://127.0.0.1:8080/configs;
        proxy_set_header Host $host;
    }
}
```

生成对外订阅链接的域名通过 `VITE_BACKEND_URL` 配置（见 `.env.example`），生产环境请设为后端实际对外域名。

## 后端改动说明

后端原有接口没有「合并配置列表」接口，前端合并配置列表页依赖它，因此给后端补充了 **一个** 接口：

- `GET /clash_configs_merge` —— 返回所有合并配置（REST 风格：列表用资源根路径）

对应改动：`ClashConfigsMergeService` / `ClashConfigsMergeServiceImpl` 增加 `list()`，`ClashConfigsMergeController` 增加 `@GetMapping`。

## 已知后端行为

- `GET /clash_configs` 只返回 `enabled = true` 的订阅，停用后将从列表消失（后端 `Example.of(ClashConfig(enabled = true))` 过滤所致）。
- 合并配置后端未提供删除接口。
