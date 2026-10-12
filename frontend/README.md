# clash-configs-web

`clash-configs` Go 后端的前端管理界面，基于 **Vite + Vue 3 + TypeScript + shadcn-vue + Tailwind CSS**。

## 功能

- 登录（session cookie）
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

开发服务器运行在 `http://localhost:5173`，通过 Vite 代理把 `/api/*` 转发到 Go 后端（默认 `https://config.jijuzhilian.com`，可用 `PROXY_TARGET` 覆盖；去掉 `/api` 前缀）。

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

生成的订阅链接使用当前页面的协议和域名，因此部署时页面域名应与对外订阅服务域名一致。

Go 后端 API 的详细列表见项目根目录 [README](../README.md#api-reference)。
