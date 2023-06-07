# goPandora

使用golang重构的Pandora-Cloud服务程序

Pandora项目地址：https://github.com/pengzhile/pandora



# 使用

1. 从[Releases](https://github.com/HamsterAPig/goPandora/releases)下载最新的版本并解压到合适的目录
2. 执行可执行文件`goPandora`



## 命令行参数

* -s: 服务器监听地址，可以是"127.0.0.1:port"的形式，也可以是":port"的形式
* -p: socks代理地址
* --CHATGPT_API_PREFIX: 配置ChatGPT的代理服务器地址，eg --CHATGPT_API_PREFIX=https://ai.fakeopen.com

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
            # Show real IP in v2ray access.log
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

