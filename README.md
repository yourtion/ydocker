# ydocker

《自己动手写Docker》实践与总结，针对代码功能与结构进行调整与优化，同时添加相应的注释与测试代码。

## 使用

### 构建

```shell
$ go build
```

### 运行

```shell
$ ./ydocker run -ti busybox sh
$ ./ydocker network create --subnet 10.0.1.0/24 --driver bridge test_bridge
$ ./ydocker network list
$ ./ydocker network remove test_bridge
$ ./ydocker run -ti -p 8080:8080 -net test_bridge --name demo busybox top
$ ./ydocker stop demo
$ ./ydocker rm demo
```

### 测试

```shell
$ sh run_test.sh
$ sh run_test_network.sh
```

## TODO

- [ ] volume 参数支持多个
- [ ] 支持自定义运行路径（当前为`/root`）
- [ ] 数据文件存取加锁
- [ ] 清理 iptables 中的 portMapping 配置
- [ ] 检测已经存在的 port 冲突
- [ ] 实现 image 相关功能
