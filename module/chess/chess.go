package chess

import (
	"regexp"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/aimerneige/MiraiChess/module/chess/database"
	"github.com/aimerneige/MiraiChess/module/chess/service"
	"gopkg.in/yaml.v2"
)

var instance *chess
var logger = utils.GetModuleLogger("aimerneige.chess")

var chessConfig struct {
	Disallowed []int64 `json:"disallowed"`
	BlackList  []int64 `json:"blacklist"`
	ELO        struct {
		Enable  bool `json:"enable"`
		Default int  `json:"default"`
		DB      struct {
			Type  string `json:"type"`
			MySQL struct {
				Username string `json:"username"`
				Password string `json:"password"`
				Host     string `json:"host"`
				Port     string `json:"port"`
				Database string `json:"database"`
				Charset  string `json:"charset"`
			} `json:"mysql"`
			SQLite struct {
				Path string `json:"path"`
			} `json:"sqlite"`
		}
	} `json:"elo"`
}

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
	path := config.GlobalConfig.GetString("aimerneige.chess.path")
	if path == "" {
		path = "./chess.yaml"
	}
	bytes := utils.ReadFile(path)
	if err := yaml.Unmarshal(bytes, &chessConfig); err != nil {
		logger.WithError(err).Errorf("Unable to read config file in %s", path)
	}
	service.UpdateELOConfig(chessConfig.ELO.Enable, chessConfig.ELO.Default)
}

// PostInit 第二次初始化
// 再次过程中可以进行跨 Module 的动作
// 如通用数据库等等
func (c *chess) PostInit() {
	if chessConfig.ELO.Enable {
		databaseType := chessConfig.ELO.DB.Type
		switch databaseType {
		case "mysql":
			mysqlDatabase := database.MysqlDatabase{
				UserName: chessConfig.ELO.DB.MySQL.Username,
				Password: chessConfig.ELO.DB.MySQL.Password,
				Host:     chessConfig.ELO.DB.MySQL.Host,
				Port:     chessConfig.ELO.DB.MySQL.Port,
				Database: chessConfig.ELO.DB.MySQL.Database,
				CharSet:  chessConfig.ELO.DB.MySQL.Charset,
			}
			database.InitDatabase(mysqlDatabase)
			logger.Info("Init mysql database success ", mysqlDatabase)
		case "sqlite":
			sqliteDatabase := database.SqliteDatabase{
				FilePath: chessConfig.ELO.DB.SQLite.Path,
			}
			database.InitDatabase(sqliteDatabase)
			logger.Info("Init sqlite database success ", sqliteDatabase)
		default:
			logger.Fatal("Unsupported database type: " + databaseType)
		}
	}
}

// Serve 注册服务函数部分
func (c *chess) Serve(b *bot.Bot) {
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		// 判断是否在黑名单中，如果是则忽略消息
		if inBlacklist(msg.Sender.Uin) {
			return
		}
		// 过滤消息来源，如果在禁用列表中则忽略消息
		if isDisallowedGroupCode(msg.GroupCode) {
			return
		}
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
			replyMsg = service.Resign(msg.GroupCode, msg.Sender, logger)
		case msgString == "draw" || msgString == "和棋":
			replyMsg = service.Draw(msg.GroupCode, msg.Sender, logger)
		case msgString == "abort" || msgString == "中断":
			replyMsg = service.Abort(c, msg.GroupCode, msg.Sender, logger)
		case []rune(msgString)[0] == '!' || []rune(msgString)[0] == '！':
			moveStr := string([]rune(msgString)[1:])
			moveStr = strings.TrimSpace(moveStr)
			if !isCorrectMoveStr(moveStr) {
				return
			}
			replyMsg = service.Play(c, msg.GroupCode, msg.Sender, moveStr, logger)
		case msgString == "cheese":
			replyMsg = service.Cheese(c, msg.GroupCode, logger)
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

func isDisallowedGroupCode(grpCode int64) bool {
	for _, v := range chessConfig.Disallowed {
		if grpCode == v {
			return true
		}
	}
	return false
}

func inBlacklist(userID int64) bool {
	for _, v := range chessConfig.BlackList {
		if userID == v {
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
