package logging

import (
	"sync"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
)

var instance *logging
var logger = utils.GetModuleLogger("com.aimerneige.logging")

type logging struct {
}

func init() {
	instance = &logging{}
	bot.RegisterModule(instance)
}

func (l *logging) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "com.aimerneige.logging",
		Instance: instance,
	}
}

// Init 初始化过程
// 在此处可以进行 Module 的初始化配置
// 如配置读取
func (l *logging) Init() {
}

// PostInit 第二次初始化
// 再次过程中可以进行跨 Module 的动作
// 如通用数据库等等
func (l *logging) PostInit() {
}

// Serve 注册服务函数部分
func (l *logging) Serve(b *bot.Bot) {
	registerLog(b)
}

// Start 此函数会新开携程进行调用
// ```go
//
//	go exampleModule.Start()
//
// ```
// 可以利用此部分进行后台操作
// 如 http 服务器等等
func (l *logging) Start(b *bot.Bot) {
}

// Stop 结束部分
// 一般调用此函数时，程序接收到 os.Interrupt 信号
// 即将退出
// 在此处应该释放相应的资源或者对状态进行保存
func (l *logging) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
}

func logGroupMessage(msg *message.GroupMessage) {
	logger.
		WithField("from", "GroupMessage").
		WithField("MessageID", msg.Id).
		WithField("MessageIID", msg.InternalId).
		WithField("GroupCode", msg.GroupCode).
		WithField("SenderID", msg.Sender.Uin).
		Info(msg.ToString())
}

func logPrivateMessage(msg *message.PrivateMessage) {
	logger.
		WithField("from", "PrivateMessage").
		WithField("MessageID", msg.Id).
		WithField("MessageIID", msg.InternalId).
		WithField("SenderID", msg.Sender.Uin).
		WithField("Target", msg.Target).
		Info(msg.ToString())
}

func logFriendMessageRecallEvent(event *client.FriendMessageRecalledEvent) {
	logger.
		WithField("from", "FriendsMessageRecall").
		WithField("MessageID", event.MessageId).
		WithField("SenderID", event.FriendUin).
		Info("friend message recall")
}

func logGroupMessageRecallEvent(event *client.GroupMessageRecalledEvent) {
	logger.
		WithField("from", "GroupMessageRecall").
		WithField("MessageID", event.MessageId).
		WithField("GroupCode", event.GroupCode).
		WithField("SenderID", event.AuthorUin).
		WithField("OperatorID", event.OperatorUin).
		Info("group message recall")
}

func logGroupMuteEvent(event *client.GroupMuteEvent) {
	logger.
		WithField("from", "GroupMute").
		WithField("GroupCode", event.GroupCode).
		WithField("OperatorID", event.OperatorUin).
		WithField("TargetID", event.TargetUin).
		WithField("MuteTime", event.Time).
		Info("group mute")
}

func logDisconnect(event *client.ClientDisconnectedEvent) {
	logger.
		WithField("from", "Disconnected").
		WithField("reason", event.Message).
		Warn("bot disconnected")
}

func registerLog(b *bot.Bot) {
	b.GroupMessageRecalledEvent.Subscribe(func(qqClient *client.QQClient, event *client.GroupMessageRecalledEvent) {
		logGroupMessageRecallEvent(event)
	})

	b.GroupMessageEvent.Subscribe(func(qqClient *client.QQClient, groupMessage *message.GroupMessage) {
		logGroupMessage(groupMessage)
	})

	b.GroupMuteEvent.Subscribe(func(qqClient *client.QQClient, event *client.GroupMuteEvent) {
		logGroupMuteEvent(event)
	})

	b.PrivateMessageEvent.Subscribe(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		logPrivateMessage(privateMessage)
	})

	b.FriendMessageRecalledEvent.Subscribe(func(qqClient *client.QQClient, event *client.FriendMessageRecalledEvent) {
		logFriendMessageRecallEvent(event)
	})

	b.DisconnectedEvent.Subscribe(func(qqClient *client.QQClient, event *client.ClientDisconnectedEvent) {
		logDisconnect(event)
	})
}
