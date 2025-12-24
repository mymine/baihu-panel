
轻量级定时任务管理系统，基于 Go + Vue 3 构建。

> ⚠️ 出于安全考虑，目前仅支持 Docker 部署方式。[镜像地址](https://github.com/engigu/BaiHu/pkgs/container/baihu)

## 快速部署

```bash
docker run -d \
  --name baihu \
  -p 8052:8052 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/envs:/app/envs \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  ghcr.io/engigu/baihu:latest
```

启动后访问：http://localhost:8052

默认账号：`admin` / `123456`


