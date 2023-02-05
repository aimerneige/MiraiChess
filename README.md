# MiraiChess

![MiraiChess](https://socialify.git.ci/aimerneige/MiraiChess/image?description=1&font=Bitter&forks=1&issues=1&owner=1&pattern=Circuit%20Board&pulls=1&stargazers=1&theme=Light)

> 本来想起名为「MiraiGoChess」的，意为 Mirai 框架 + Go 语言 + 国际象棋，但是社区内有一个项目 [MiraiGoChess](https://github.com/Minxyzgo/MiraiGoChess) 已经用了这个名字，所以咱就叫 「MiraiChess」了。

> **Note**\
> 目前的代码非常屎，仅供参考，不要学。\
> 因为刚开始功能比较少就偷懒了，后面加了好多功能，懒得重构就越写越屎了。<sub>~能跑就行了要什么可扩展性~</sub>\
> 这个项目的功能已经很完善了，后续也不会有大功能更新，已经没必要花时间重构。
> 如果您愿意花时间重构这个项目，欢迎 pr。

使用 [MiraiGo-Template](https://github.com/Logiase/MiraiGo-Template) 实现的国际象棋机器人。

<img src="https://raw.githubusercontent.com/aimerneige/MiraiChess/master/img/bot.jpeg" alt="bot">

## Language

I assume that most of the users of this repo are Chinese, so there will be no English support for README and documents. If you need, post a issue.

## 模块化支持

如果你希望以模块化的方式将本 bot 集成到已有项目中，可以使用如下仓库：

[MiraiGo-module-chess](https://github.com/yukichan-bot-module/MiraiGo-module-chess)

## 如何使用

### 安装

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
mkdir -p /opt/MiraiChess
tar -xzvf MiraiChess-linux-amd64-v1.9.0.tar.gz -C /opt/MiraiChess
cd /opt/MiraiChess
```

按需适当修改 `./config` 下配置文件（默认配置可以直接运行）：

```
vim ./config/application.yaml
vim ./config/chess.go
```

运行启动脚本，下载依赖并确保机器人正常运行（必需至少运行一次启动脚本）：

```bash
./start.sh
```

机器人运行无误后，可以安装 service 文件：

```bash
cp ./mirai-chess.service /etc/systemd/system/mirai-chess.service
```

启动机器人：

```bash
systemctl start mirai-chess.service
```

开机自启：

```bash
systemctl enable mirai-chess.service
```

### 升级

下载解压 [release](https://github.com/aimerneige/MiraiChess/releases) 后直接执行 `update.sh` 即可。

:warning: 升级脚本只简单地执行了如下任务：

1. 关闭服务
2. 替换新的可执行文件
3. 替换 README 文件
4. 重启服务

更新脚本不会修改配置文件，当有功能更新时请注意手动更改配置文件。

请不要依赖这个脚本，如果遇到问题请备份配置文件后重新安装。

## 如何编译

> **Warning**\
> 如果你打算手动编译，请在 [release](https://github.com/aimerneige/MiraiChess/releases) 下载一份最新稳定代码，不要直接使用 master 分支。

环境依赖：

1. [golang](https://go.dev/dl/)
2. [python](https://www.python.org/downloads/)
3. [make](https://www.gnu.org/software/make/)

编译可执行文件：

```bash
make build
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
本项目主要是希望提供一个在群内下棋的环境，重要的是大家一起围观、交流和讨论棋局，而不是单纯实现对局。太多的对局同时进行不仅会导致群消息过多炸群，而且也不利于交流。如果只是需要下棋，chess.com 的邀请链接完全可以创建无限的棋局。<sub>~绝对不是开发者懒得写~</sub>

### 扫码登录时被风控如何处理

如果服务器网络环境被风控，在本地执行后将生成的 `device.json` 及 `session.token` 上传至服务器即可。

#### ~~本地也跑不起来怎么办~~

可以尝试换个帐号，腾讯的风控很玄学的，本项目也没有什么好的办法。

以下是个人总结的一些玄学方法：

1. 尽量不要使用刚注册的小号
2. 帐号要实名并绑定手机号
3. 开启设备锁、人脸识别等安全工具
4. 不要频繁切换登录 IP
5. 帐号多加点好友和群

### 是否支持 Windows

本项目只建议使用 Linux 服务器运行本项目，如果你一定要使用 Windows，请按照下面的方法安装使用：

#### 安装依赖

1. [inkscape](https://inkscape.org/release/)
2. [python](https://www.python.org/downloads/)
3. [python-chess](https://github.com/niklasf/python-chess)

#### 编译

将本项目 clone 之后，在项目路径下执行如下指令：

```bash
go build -o bin/device.exe cmd/device/device.go
go build -o bin/bot.exe cmd/bot/bot.go
```

注意，编译本项目需要安装配置 Go 和 MinGW

#### 修改配置文件

在 `config/chess.yaml` 文件中修改 inkscape 可执行文件路径为安装路径。

#### 生成设备文件

在项目路径下执行 `device.exe`

```bat
.\bin\device.exe
```

#### 启动机器人

```bat
.\bin\bot.exe
```

## 交流群

点击链接或扫码加入 QQ 群:

[857066811](https://qm.qq.com/cgi-bin/qm/qr?k=rMtw1SlmoFOp08i5Zw5bM361ljIyzVA-&authKey=9OUzro5oH5CnnFaAbIMwa60987+8ZMwu5GvUAlFUzDIQKVL91z9zUhWp6m1Kayf8&noverify=0)

![qrcode 857066811](img/qr-code.png)

## LICENSE

<a href="https://www.gnu.org/licenses/agpl-3.0.en.html">
<img src="https://www.gnu.org/graphics/agplv3-155x51.png">
</a>

本项目使用 `AGPLv3` 协议开源，您可以在 [GitHub](https://github.com/aimerneige/MiraiChess) 获取本项目源代码。为了整个社区的良性发展，我们强烈建议您做到以下几点：

- **间接接触（包括但不限于使用 `Http API` 或 跨进程技术）到本项目的软件使用 `AGPLv3` 开源**
- **不鼓励，不支持一切商业使用**

## 开源相关

- [MiraiGo-Template](https://github.com/Logiase/MiraiGo-Template)
- [chess](https://github.com/notnil/chess)
- [python-chess](https://github.com/niklasf/python-chess)
