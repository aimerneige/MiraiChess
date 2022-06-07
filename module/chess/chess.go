package chess

import (
	"regexp"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/aimerneige/MiraiChess/module/chess/service"
)

var instance *chess
var logger = utils.GetModuleLogger("internal.logging")
var allowedGroup []int64

type chess struct {
}

func init() {
	instance = &chess{}
	bot.RegisterModule(instance)
}

func (c *chess) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "aimerneige.chess",
		Instance: instance,
	}
}

// Init 初始化过程
// 在此处可以进行 Module 的初始化配置
// 如配置读取
func (c *chess) Init() {
	// 读取配置文件并初始化 allowedGroup
	// 咕咕咕
}

// PostInit 第二次初始化
// 再次过程中可以进行跨 Module 的动作
// 如通用数据库等等
func (c *chess) PostInit() {
}

// Serve 注册服务函数部分
func (c *chess) Serve(b *bot.Bot) {
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		// 过滤消息来源，仅对配置文件中指定的群提供服务
		// if !isAllowedGroupCode(msg.GroupCode) {
		// 	return
		// }
		// 忽略匿名消息
		if msg.Sender.IsAnonymous() {
			return
		}
		// 解析消息内容
		var replyMsg *message.SendingMessage
		switch msgString := msg.ToString(); {
		case msgString == "chess" || msgString == "下棋":
			replyMsg = service.Game(c, msg.GroupCode, msg.Sender, logger)
		case msgString == "resign" || msgString == "认输":
			replyMsg = service.Resign(msg.GroupCode, msg.Sender)
		case msgString == "draw" || msgString == "和棋":
			replyMsg = service.Draw(msg.GroupCode, msg.Sender)
		case []rune(msgString)[0] == '!' || []rune(msgString)[0] == '！':
			moveStr := string([]rune(msgString)[1:])
			if !isCorrectMoveStr(moveStr) {
				return
			}
			replyMsg = service.Play(c, msg.GroupCode, msg.Sender, moveStr, logger)
		default:
			return
		}
		if replyMsg != nil {
			c.SendGroupMessage(msg.GroupCode, replyMsg)
		}
	})
}

// Start 此函数会新开携程进行调用
// ```go
// 		go exampleModule.Start()
// ```
// 可以利用此部分进行后台操作
// 如 http 服务器等等
func (c *chess) Start(b *bot.Bot) {
}

// Stop 结束部分
// 一般调用此函数时，程序接收到 os.Interrupt 信号
// 即将退出
// 在此处应该释放相应的资源或者对状态进行保存
func (c *chess) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
}

func isAllowedGroupCode(grpCode int64) bool {
	for _, v := range allowedGroup {
		if grpCode == v {
			return true
		}
	}
	return false
}

func isCorrectMoveStr(moveStr string) bool {
	if len(moveStr) == 0 {
		return false
	}
	const PATTERN = "([0-9]|[A-Z]|[a-z])+"
	reg := regexp.MustCompile(moveStr)
	return reg.MatchString(moveStr)
}
