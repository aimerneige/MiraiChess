# MiraiChess

> 本来想起名为「MiraiGoChess」的，意为 Mirai 框架 + Go 语言 + 国际象棋，但是社区内有一个项目 [MiraiGoChess](https://github.com/Minxyzgo/MiraiGoChess) 已经用了这个名字，所以咱就叫 「MiraiChess」了。

使用 [MiraiGo-Template](https://github.com/Logiase/MiraiGo-Template) 实现的国际象棋机器人。

## Language

I assume that most of the users of this repo are Chinese, so there will be no English support for README and documents. If you need, post a issue.

## 如何使用

准备如下环境：

1. Linux 服务器 （笔记本啥的也可以，关键要有 Linux 环境）
2. [golang](https://go.dev/dl/)
3. [make](https://www.gnu.org/software/make/)

将项目下载到本地：

```bash
git clone https://github.com/aimerneige/MiraiChess.git
cd MiraiChess
```

执行脚本下载 inkscape：

```bash
./scripts/download_inkscape.sh
```

编译项目：

```bash
make all
```

完成后在 `./bin` 目录下会看到可执行文件。

启动之前，先生成设备文件，并将生成的 `device.json` 移动到项目根目录下：

```bash
cd test
go test
mv device.json ../
```

适当修改 `./config` 下配置文件

启动机器人

```bash
./bin/mirai-chess-bot-linux-amd64-v0.0.1
```

测试无误后可将其转为后台执行：

```bash
nohup ./bin/mirai-chess-bot-linux-amd64-v0.0.1 &
```

## TODO

- [ ] 提供 service 文件
- [ ] 提供 docker 支持
- [ ] 完善文档

## LICENSE

<a href="https://www.gnu.org/licenses/agpl-3.0.en.html">
<img src="https://www.gnu.org/graphics/agplv3-155x51.png">
</a>
