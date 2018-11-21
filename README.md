# forward-proxy

* 监听端口 8888
* 参数 url、key（可选 AES 密钥）
* 仅支持 GET

```
http://localhost:8888?url=https://www.google.com
```

response 返回 JSON 格式数据：

```json
{
    "html": "<html>...</html>",
    "headers": [
        {"content-type": "text/html;charset=utf-8"},
        {"xxx": "..."}
    ]
}
```
