# forward-proxy

使用 POST 提交代理请求：

```json
{
    "url": "https://github.com", 
    "headers": [
        {
            "User-Agent": ""
        }, 
        {
            "Connection": "close"
        }
    ], 
    "method": "get"
}
```

响应返回 JSON 格式数据：

```json
{
    "status": 200,
    "contentType": "",
    "body": "...",
    "headers": [
        {"content-type": "text/html;charset=utf-8"},
        {"xxx": "..."}
    ]
}
```
