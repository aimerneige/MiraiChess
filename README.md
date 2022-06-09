# MiraiChess

> 本来想起名为「MiraiGoChess」的，意为 Mirai 框架 + Go 语言 + 国际象棋，但是社区内有一个项目 [MiraiGoChess](https://github.com/Minxyzgo/MiraiGoChess) 已经用了这个名字，所以咱就叫 「MiraiChess」了。

使用 [MiraiGo-Template](https://github.com/Logiase/MiraiGo-Template) 实现的国际象棋机器人。

<img src="https://raw.githubusercontent.com/aimerneige/MiraiChess/master/img/chess-girl.jpg" alt="anime girl" width="360" height="480">

图源：<https://www.pixiv.net/en/artworks/90731304>

## Language

I assume that most of the users of this repo are Chinese, so there will be no English support for README and documents. If you need, post a issue.

## 如何使用

准备如下环境：

1. Linux 服务器 （笔记本啥的也可以，关键要有 Linux 环境）
2. [golang](https://go.dev/dl/)
3. [python](https://www.python.org/downloads/)
4. [make](https://www.gnu.org/software/make/)

将项目下载到本地：

```bash
git clone https://github.com/aimerneige/MiraiChess.git
cd MiraiChess
```

执行脚本下载 inkscape：\
:warning: 脚本中使用了相对路径，请务必在项目根目录下执行脚本。

```bash
./scripts/download_inkscape.sh
```

安装 python 库 [python-chess](https://github.com/niklasf/python-chess)：

```bash
pip install python-chess
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

## Q&A

### 是否会支持群内多盘对局同时进行

每个群内同时只能存在一盘对局，如果有多盘对局同时进行的需求可以 fork 之后自己改。\
本项目主要是希望提供一个在群内下棋的环境，重要的是大家一起围观、交流和讨论棋局，而不是单纯实现对局。太多的对局同时进行不仅会导致群消息过多炸群，而且也不利于交流。如果只是需要下棋，chess.com 的邀请链接完全可以创建无限的棋局。<sub>~绝对不是开发者懒得写！！！~</sub>

### 扫码登录时被风控如何处理

如果服务器网络环境被风控，在本地执行后将生成的 `device.json` 及 `session.token` 上传至服务器即可。

### 是否支持 Windows

项目代码本身可以编译 Windows 版本，但在 svg 转 png 时用到了 [inkscape](https://inkscape.org/)，该软件提供 Windows 版本但本项目没有测试其可用性，如果您有这方面的需求，可以尝试在 Windows 下调用 inkscape 或重写 svg 转 png 的相关代码。（PR WELCOME）

## TODO

- [ ] 提供 service 文件
- [ ] 提供 docker 支持
- [ ] 对局分析
- [x] 完善文档

## LICENSE

<a href="https://www.gnu.org/licenses/agpl-3.0.en.html">
<img src="https://www.gnu.org/graphics/agplv3-155x51.png">
</a>
