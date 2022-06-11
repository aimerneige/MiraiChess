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
2. [python](https://www.python.org/downloads/)

安装 python 库 [python-chess](https://github.com/niklasf/python-chess)：

```bash
pip install python-chess
```

下载最新 [release](https://github.com/aimerneige/MiraiChess/releases) 或 [自己编译](https://github.com/aimerneige/MiraiChess#%E5%A6%82%E4%BD%95%E7%BC%96%E8%AF%91)。

解压 release：

```bash
tar -xzvf MiraiChess-linux-amd64-v1.0.0.tar.gz
```

适当修改 `./config` 下配置文件。

启动项目：

```bash
./start.sh
```

## 如何编译

环境依赖：

1. [golang](https://go.dev/dl/)
2. [python](https://www.python.org/downloads/)
3. [make](https://www.gnu.org/software/make/)

编译可执行文件：

```bash
make all
```

发布 release：

```bash
make release
```

测试运行：

```bash
make run
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
- [x] 完善文档

## LICENSE

<a href="https://www.gnu.org/licenses/agpl-3.0.en.html">
<img src="https://www.gnu.org/graphics/agplv3-155x51.png">
</a>

本项目使用 `AGPLv3` 协议开源，您可以在 [GitHub](https://github.com/aimerneige/MiraiChess) 获取本项目源代码。为了整个社区的良性发展，我们强烈建议您做到以下几点：

- **间接接触（包括但不限于使用 `Http API` 或 跨进程技术）到本项目的软件使用 `AGPLv3` 开源**
- **不鼓励，不支持一切商业使用**

## 开源相关

- [MiraiGo-Template](https://github.com/Logiase/MiraiGo-Template)
- [MiraiGo-module-autoreply](https://github.com/Logiase/MiraiGo-module-autoreply)
- [chess](https://github.com/notnil/chess)
- [python-chess](https://github.com/niklasf/python-chess)
