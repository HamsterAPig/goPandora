# goPandora

使用golang重构的Pandora-Cloud服务程序

Pandora项目地址：https://github.com/pengzhile/pandora



# 使用

1. 从[Releases](https://github.com/HamsterAPig/goPandora/releases)下载最新的版本并解压到合适的目录
2. 执行可执行文件`goPandora`



## 配置选项

### main
- `listen` (字符串): 监听地址。
- `debug-level` (字符串): 日志等级。
- `ChatGPT_API_PREFIX` (字符串): GhatGPT网址前缀。
- `endpoint` (字符串): 后端服务器地址。
- `enable-verify-share-page` (布尔值): 是否启用分享页验证。
- `enable-day-api-prefix` (布尔值): 启用日抛域名支持。

### cloudflare
> 启用在有IP触发404之后，自动在cloudflare上面将IP拉黑

- `email`: cloudflare账号的邮箱
- `api_key`: api_key，在cloudflare官网进入网站概述页面的右下角
- `zone_id`: 同上
- `notes`: 添加block规则时候的备注

```yml
main:
  listen: ":8080"
  debug-level: "info"
  ChatGPT_API_PREFIX: "https://ai.fakeopen.com"
  endpoint: "http://127.0.0.1:8899"
  enable-verify-share-page: true
  enable-day-api-prefix: true
cloudflare:
  email: your cloudflare email
  api_key: your api_key
  zone_id: your zone_id
  notes: blocked ip from goPandora
```



## Nginx 前置代理配置

本项目中并没有内置https的处理，所以还得配置Nginx作为前置代理服务器，需要将以下内容加入Nginx的配置文件当中

```conf
        location / {
            proxy_buffering off;
            proxy_cache off;
            proxy_redirect off;
            proxy_pass http://127.0.0.1:2346/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        location ~ .* {
            proxy_buffering off;
            proxy_cache off;
            proxy_pass http://127.0.0.1:2346;
            proxy_set_header Host $http_host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
```

