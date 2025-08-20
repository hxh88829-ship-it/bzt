# bzt
## Docker
```bash
# build
docker build -t valueguard .

# run
docker run -d --name bzt -p 8000:8000 -t valueguard

# 此服务需要开放外网端口
需要公网固定IP，暴露对外指定端口：8000


```

