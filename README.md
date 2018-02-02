# containerd 学习

### 准备

修改 var/run/docker/containerd/containerd.toml

```
uid 为当前用户的uid
gid 为当前用户的gid
```


### proxy sock to http server

```
# 代理debug服务到http://localhost:8080 访问
make run 
```

