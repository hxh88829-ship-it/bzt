# bzt
## Docker
```bash
# build
docker build -t valueguard .

# run
docker run -d --name bzt -p 8000:8000 -t valueguard

# 此服务需要开放外网端口
需要公网固定IP，暴露对外指定端口：8000


# 配置机器安全组访问环境变量
Apikey     (签名服务的apikey)
BaseUrl    (签名服务的ip)
KeyId      (新建一个保值通项目的kms keyid)
OwnerAddress  (keyid对应的钱包地址)
HmacKey    (hash消息认证码)
RpcUrl     (DTC 节点 RPC)
```

