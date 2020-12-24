#!/bin/bash

#set -e

go build

run () {
    echo $1 ": \"${@:2}\""
    echo "--------------"
    ${@:2}
    echo "--------------"
    echo ""
}


ip link delete dev test_bridge

run "调用 ydocker 创建网络" ./ydocker network create --subnet 10.0.1.0/24 --driver bridge test_bridge
run "调用 ydocker 列举创建的网络" ./ydocker network list
run "查看创建出的 Bridge 设备" ip link show dev test_bridge
run "查看地址配置" ip link show dev test_bridge
run "查看路由配置" ip route show dev test_bridge
run "查看 iptables 配置的 MASQUERADE 规则" iptables -t nat -vnL POSTROUTING

run "调用 ydocker 删除刚才创建的网络" ./ydocker network remove test_bridge
run "列表中已经看不到网络" ./ydocker network list
run "看到网络对应的设备也被删除" ip link show dev test_bridge