# 1.gw node
### 1.build
```shell
go build -x github.com/ds/depaas/cmd/gw
```
### run
```shell
./gw
```
# 2.miner node
### 1.build
```shell
go build -x -tags main github.com/ds/depaas/cmd/miner
```
### run
```shell
./miner -data /mnt -size 100GB 
```
# Architecture
![Architecture](./docs/me.svg)