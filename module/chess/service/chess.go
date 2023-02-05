package service

import (
	"bytes"
	_ "embed" // embed for cheese
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/aimerneige/MiraiChess/module/chess/database"
	"github.com/notnil/chess"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//go:embed assets/cheese.jpeg
var cheeseData []byte

//go:embed assets/help.txt
var helpString string

//go:embed scripts/board2svg.py
var pythonScriptBoard2SVG string

//go:embed scripts/pgn2gif.py
var pythonScriptPGN2GIF string

var instance *chessService
var inkscapePath string
var tempFileDir string
var eloEnabled bool
var eloDefault int
var boardTheme string

type chessService struct {
	gameRooms map[int64]chessRoom
}

type chessRoom struct {
	chessGame    *chess.Game
	whitePlayer  int64
	whiteName    string
	blackPlayer  int64
	blackName    string
	drawPlayer   int64
	lastMoveTime int64
}

func init() {
	instance = &chessService{
		gameRooms: make(map[int64]chessRoom, 1),
	}
}

// UpdateFSConfig update fs config
func UpdateFSConfig(inkscape, temp string) {
	inkscapePath = inkscape
	tempFileDir = temp
}

// UpdateELOConfig update elo config
func UpdateELOConfig(enabled bool, defaultValue int) {
	eloEnabled = enabled
	eloDefault = defaultValue
}

// UpdateBoardTheme update board theme config
func UpdateBoardTheme(theme string) {
	boardTheme = theme
	if boardTheme == "" {
		boardTheme = "default"
	}
}

// Game 下棋
func Game(c *client.QQClient, groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		if room.blackPlayer != 0 {
			// 检测对局是否已存在超过 6 小时
			if (time.Now().Unix() - room.lastMoveTime) > 21600 {
				return abortGame(groupCode, "对局已存在超过 6 小时，游戏结束。", logger).Append(message.NewText("\n\n已有对局已被中断，如需创建新对局请重新发送指令。")).Append(message.NewAt(sender.Uin))
			}
			// 对局在进行
			msg := textWithAt(sender.Uin, "对局已在进行中，无法创建或加入对局，当前对局玩家为：")
			if room.whitePlayer != 0 {
				msg.Append(message.NewAt(room.whitePlayer))
			}
			if room.blackPlayer != 0 {
				msg.Append(message.NewAt(room.blackPlayer))
			}
			msg.Append(message.NewText("，群主或管理员发送「中断」或「abort」可中断对局（自动判和）。"))
			return msg
		}
		if sender.Uin == room.whitePlayer {
			return textWithAt(sender.Uin, "请等候其他玩家加入游戏。")
		}
		room.blackPlayer = sender.Uin
		room.blackName = sender.Nickname
		instance.gameRooms[groupCode] = room
		boardImgEle, ok, errMsg := getBoardElement(c, groupCode, logger)
		if !ok {
			return errorText(errMsg)
		}
		return simpleText("黑棋已加入对局，请白方下棋。").Append(message.NewAt(room.whitePlayer)).Append(boardImgEle)
	}
	instance.gameRooms[groupCode] = chessRoom{
		chessGame:    chess.NewGame(),
		whitePlayer:  sender.Uin,
		whiteName:    sender.Nickname,
		blackPlayer:  0,
		blackName:    "",
		drawPlayer:   0,
		lastMoveTime: time.Now().Unix(),
	}
	return simpleText("已创建新的对局，发送「下棋」或「chess」可加入对局。")
}

// Abort 中断对局
func Abort(c *client.QQClient, groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	// 判断是否是群主或管理员
	groupMemberInfo, err := c.GetMemberInfo(groupCode, sender.Uin)
	if err != nil {
		logger.WithError(err).Errorf("Fail to get group member info.")
		return nil
	}
	// 不是管理员，忽略消息
	if groupMemberInfo.Permission != client.Administrator && groupMemberInfo.Permission != client.Owner {
		return nil
	}
	// 检查并处理对局
	if _, ok := instance.gameRooms[groupCode]; ok {
		return abortGame(groupCode, "对局已被管理员中断，游戏结束。", logger)
	}
	return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
}

// Draw 和棋
func Draw(c *client.QQClient, groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		if sender.Uin == room.whitePlayer || sender.Uin == room.blackPlayer {
			room.lastMoveTime = time.Now().Unix()
			if room.drawPlayer == 0 {
				room.drawPlayer = sender.Uin
				instance.gameRooms[groupCode] = room
				return textWithAt(sender.Uin, "请求和棋，发送「和棋」或「draw」接受和棋。走棋视为拒绝和棋。")
			}
			if room.drawPlayer == sender.Uin {
				return textWithAt(sender.Uin, "已发起和棋请求，请勿重复发送。")
			}
			room.chessGame.Draw(chess.DrawOffer)
			chessString := getChessString(room)
			eloString := ""
			if eloEnabled && len(room.chessGame.Moves()) > 4 {
				dbService := NewDBService(database.GetDB())
				if err := dbService.CreatePGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
					logger.WithError(err).Error("Fail to create PGN.")
				}
				whiteScore, blackScore := 0.5, 0.5
				elo, err := getELOString(room, whiteScore, blackScore, dbService)
				if err != nil {
					logger.WithError(err).Error("Fail to get eloString. " + eloString)
				}
				eloString = elo
			}
			replyMsg := textWithAt(sender.Uin, "接受和棋，游戏结束。\n"+eloString+chessString)
			gif, msg, err := generateGIF(c, groupCode, chessString, logger)
			if err != nil {
				logger.WithError(err).Error("Fail to generate GIF.")
				replyMsg.Append(message.NewText("\n\n[GIF - ERROR]\n" + msg))
			} else {
				replyMsg.Append(gif)
			}
			delete(instance.gameRooms, groupCode)
			return replyMsg
		}
		return textWithAt(sender.Uin, "不是对局中的玩家，无法请求和棋。")
	}
	return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
}

// Resign 认输
func Resign(c *client.QQClient, groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		// 检查是否是当前游戏玩家
		if sender.Uin == room.whitePlayer || sender.Uin == room.blackPlayer {
			// 如果对局未建立，中断对局
			if room.whitePlayer == 0 || room.blackPlayer == 0 {
				delete(instance.gameRooms, groupCode)
				return simpleText("对局已释放。")
			}
			var resignColor chess.Color
			if sender.Uin == room.whitePlayer {
				resignColor = chess.White
			} else {
				resignColor = chess.Black
			}
			if isAprilFoolsDay() {
				if resignColor == chess.White {
					resignColor = chess.Black
				} else {
					resignColor = chess.White
				}
			}
			room.chessGame.Resign(resignColor)
			chessString := getChessString(room)
			eloString := ""
			if eloEnabled && len(room.chessGame.Moves()) > 4 {
				dbService := NewDBService(database.GetDB())
				if err := dbService.CreatePGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
					logger.WithError(err).Error("Fail to create PGN.")
				}
				whiteScore, blackScore := 1.0, 1.0
				if resignColor == chess.White {
					whiteScore = 0.0
				} else {
					blackScore = 0.0
				}
				elo, err := getELOString(room, whiteScore, blackScore, dbService)
				if err != nil {
					logger.WithError(err).Error("Fail to get eloString. " + eloString)
				}
				eloString = elo
			}
			replyMsg := textWithAt(sender.Uin, "认输，游戏结束。\n"+eloString+chessString)
			if isAprilFoolsDay() {
				replyMsg = textWithAt(sender.Uin, "对手认输，游戏结束，你胜利了。\n"+eloString+chessString)
			}
			gif, msg, err := generateGIF(c, groupCode, chessString, logger)
			if err != nil {
				logger.WithError(err).Error("Fail to generate GIF.")
				replyMsg.Append(message.NewText("\n\n[GIF - ERROR]\n" + msg))
			} else {
				replyMsg.Append(gif)
			}
			delete(instance.gameRooms, groupCode)
			return replyMsg
		}
		return textWithAt(sender.Uin, "不是对局中的玩家，无法认输。")
	}
	return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
}

// Play 走棋
func Play(c *client.QQClient, groupCode int64, sender *message.Sender, moveStr string, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		// 不是对局中的玩家，忽略消息
		if (sender.Uin != room.whitePlayer) && (sender.Uin != room.blackPlayer) && !isAprilFoolsDay() {
			return nil
		}
		// 对局未建立
		if (room.whitePlayer == 0) || (room.blackPlayer == 0) {
			return textWithAt(sender.Uin, "请等候其他玩家加入游戏。")
		}
		// 需要对手走棋
		if ((sender.Uin == room.whitePlayer) && (room.chessGame.Position().Turn() != chess.White)) || ((sender.Uin == room.blackPlayer) && (room.chessGame.Position().Turn() != chess.Black)) {
			return textWithAt(sender.Uin, "请等待对手走棋。")
		}
		room.lastMoveTime = time.Now().Unix()
		// 走棋
		if err := room.chessGame.MoveStr(moveStr); err != nil {
			return simpleText(fmt.Sprintf("移动「%s」违规，请检查，格式请参考「代数记谱法」(Algebraic notation)。", moveStr))
		}
		// 走子之后，视为拒绝和棋
		if room.drawPlayer != 0 {
			room.drawPlayer = 0
			instance.gameRooms[groupCode] = room
		}
		// 生成棋盘图片
		boardImgEle, ok, errMsg := getBoardElement(c, groupCode, logger)
		if !ok {
			return errorText(errMsg)
		}
		// 检查游戏是否结束
		if room.chessGame.Method() != chess.NoMethod {
			whiteScore, blackScore := 0.5, 0.5
			msg := "游戏结束，"
			switch room.chessGame.Method() {
			case chess.FivefoldRepetition:
				msg += "和棋，因为五次重复走子。\n"
			case chess.SeventyFiveMoveRule:
				msg += "和棋，因为七十五步规则。\n"
			case chess.InsufficientMaterial:
				msg += "和棋，因为不可能将死。\n"
			case chess.Stalemate:
				msg += "和棋，因为逼和（无子可动和棋）。\n"
			case chess.Checkmate:
				var winner string
				if room.chessGame.Position().Turn() == chess.White {
					whiteScore = 0.0
					blackScore = 1.0
					winner = "黑方"
				} else {
					whiteScore = 1.0
					blackScore = 0.0
					winner = "白方"
				}
				msg += winner
				msg += "胜利，因为将杀。\n"
			}
			chessString := getChessString(room)
			eloString := ""
			if eloEnabled && len(room.chessGame.Moves()) > 4 {
				dbService := NewDBService(database.GetDB())
				if err := dbService.CreatePGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
					logger.WithError(err).Error("Fail to create PGN.")
				}
				elo, err := getELOString(room, whiteScore, blackScore, dbService)
				if err != nil {
					logger.WithError(err).Error("Fail to get eloString. " + eloString)
				}
				eloString = elo
			}
			replyMsg := simpleText(msg + eloString + chessString).Append(boardImgEle)
			gif, msg, err := generateGIF(c, groupCode, chessString, logger)
			if err != nil {
				logger.WithError(err).Error("Fail to generate GIF.")
				replyMsg.Append(message.NewText("\n\n[GIF - ERROR]\n" + msg))
			} else {
				replyMsg.Append(gif)
			}
			delete(instance.gameRooms, groupCode)
			return replyMsg
		}
		// 提示玩家继续游戏
		var currentPlayer int64
		if room.chessGame.Position().Turn() == chess.White {
			currentPlayer = room.whitePlayer
		} else {
			currentPlayer = room.blackPlayer
		}
		return textWithAt(currentPlayer, "对手已走子，游戏继续。").Append(boardImgEle)
	}
	return textWithAt(sender.Uin, "对局不存在，发送「下棋」或「chess」可创建对局。")
}

// Cheese Easter Egg
func Cheese(c *client.QQClient, groupCode int64, logger logrus.FieldLogger) *message.SendingMessage {
	// 上传图片
	ele, err := uploadImage(c, groupCode, bytes.NewReader(cheeseData), logger)
	if err != nil {
		logger.WithError(err).Error("Unable to upload image.")
		return nil
	}
	return simpleText("Chess Cheese Cheese Chess").Append(ele)
}

// Help 帮助信息
func Help() *message.SendingMessage {
	return simpleText(helpString)
}

// Ranking 排行榜
func Ranking(c *client.QQClient, logger logrus.FieldLogger) *message.SendingMessage {
	if !eloEnabled {
		return nil
	}
	dbService := NewDBService(database.GetDB())
	ranking, err := getRankingString(dbService)
	if err != nil {
		logger.WithError(err).Errorf("Fail to get ranking string")
		return simpleText("服务器错误，无法获取排行榜信息。请联系开发者修 bug。\n反馈地址 https://github.com/aimerneige/MiraiChess/issues\n")
	}
	return simpleText(ranking)
}

// Rate 获取等级分
func Rate(c *client.QQClient, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if !eloEnabled {
		return nil
	}
	dbService := NewDBService(database.GetDB())
	rate, err := dbService.GetELORateByUin(sender.Uin)
	if err == gorm.ErrRecordNotFound {
		return simpleText("没有查找到等级分信息。请至少进行一局对局。")
	}
	if err != nil {
		logger.WithError(err).Errorf("Fail to get player rank")
		return simpleText("服务器错误，无法获取等级分信息。请联系开发者修 bug。\n反馈地址 https://github.com/aimerneige/MiraiChess/issues\n")
	}
	return simpleText(fmt.Sprintf("玩家%s目前的等级分：%d", sender.Nickname, rate))
}

func errorText(errMsg string) *message.SendingMessage {
	return simpleText("发生错误，请联系开发者修 bug。\n反馈地址 https://github.com/aimerneige/MiraiChess/issues\n错误信息：" + errMsg)
}

func simpleText(msg string) *message.SendingMessage {
	return message.NewSendingMessage().Append(message.NewText(msg))
}

func textWithAt(target int64, msg string) *message.SendingMessage {
	if target == 0 {
		return simpleText("@全体成员 " + msg)
	}
	return message.NewSendingMessage().Append(message.NewAt(target)).Append(message.NewText(msg))
}

func generateGIF(c *client.QQClient, groupCode int64, pgnStr string, logger logrus.FieldLogger) (*message.GroupImageElement, string, error) {
	if err := exec.Command("python", "-c", pythonScriptPGN2GIF, pgnStr, tempFileDir, fmt.Sprintf("%d", groupCode)).Run(); err != nil {
		logger.Info("python", " ", "-c", " ", "python_sript_pgn2gif", " ", pgnStr, " ", tempFileDir, " ", fmt.Sprintf("%d", groupCode))
		return nil, "生成 gif 时发生错误", err
	}
	gifFilePath := path.Join(tempFileDir, fmt.Sprintf("%d.gif", groupCode))
	f, err := os.Open(gifFilePath)
	if err != nil {
		logger.WithError(err).Errorf("Unable to read gif file in %s", gifFilePath)
		return nil, "读取 gif 时发生错误", err
	}
	ele, err := uploadImage(c, groupCode, f, logger)
	return ele, "", err
}

func uploadImage(c *client.QQClient, groupCode int64, img io.ReadSeeker, logger logrus.FieldLogger) (*message.GroupImageElement, error) {
	// 尝试上传图片
	ele, err := c.UploadGroupImage(groupCode, img)
	// 发生错误时重试 3 次，否则报错
	for i := 0; i < 3 && err != nil; i++ {
		ele, err = c.UploadGroupImage(groupCode, img)
	}
	if err != nil {
		logger.WithError(err).Error("Unable to upload image.")
		return nil, err
	}
	return ele, nil
}

func abortGame(groupCode int64, hint string, logger logrus.FieldLogger) *message.SendingMessage {
	room := instance.gameRooms[groupCode]
	room.chessGame.Draw(chess.DrawOffer)
	chessString := getChessString(room)
	if eloEnabled && len(room.chessGame.Moves()) > 4 {
		dbService := NewDBService(database.GetDB())
		if err := dbService.CreatePGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			logger.WithError(err).Error("Fail to create PGN.")
		}
	}
	delete(instance.gameRooms, groupCode)
	msg := simpleText(hint)
	if room.whitePlayer != 0 {
		msg.Append(message.NewAt(room.whitePlayer))
	}
	if room.blackPlayer != 0 {
		msg.Append(message.NewAt(room.blackPlayer))
	}
	msg.Append(message.NewText("\n\n" + chessString))
	return msg
}

func getBoardElement(c *client.QQClient, groupCode int64, logger logrus.FieldLogger) (*message.GroupImageElement, bool, string) {
	if room, ok := instance.gameRooms[groupCode]; ok {
		var uciStr string
		// 将最后一步走子转化为 uci 字符串
		moves := room.chessGame.Moves()
		if len(moves) != 0 {
			uciStr = moves[len(moves)-1].String()
		} else {
			uciStr = "None"
		}
		svgFilePath := path.Join(tempFileDir, fmt.Sprintf("%d.svg", groupCode))
		pngFilePath := path.Join(tempFileDir, fmt.Sprintf("%d.png", groupCode))
		// 调用 python 脚本生成 svg 文件
		if err := exec.Command("python", "-c", pythonScriptBoard2SVG, room.chessGame.FEN(), svgFilePath, uciStr, boardTheme).Run(); err != nil {
			logger.Info("python", " ", "-c", " ", "python_script_board2svg", " ", room.chessGame.FEN(), " ", svgFilePath, " ", uciStr, " ", boardTheme)
			logger.WithError(err).Error("Unable to generate svg file.")
			return nil, false, "无法生成 svg 图片"
		}
		// 调用 inkscape 将 svg 图片转化为 png 图片
		if err := exec.Command(inkscapePath, "-w", "720", "-h", "720", svgFilePath, "-o", pngFilePath).Run(); err != nil {
			logger.WithError(err).Error("Unable to convert to png.")
			return nil, false, "无法生成 png 图片"
		}
		// 尝试读取 png 图片
		f, err := os.Open(pngFilePath)
		if err != nil {
			logger.WithError(err).Errorf("Unable to read image file in %s.", pngFilePath)
			return nil, false, "无法读取 png 图片"
		}
		defer f.Close()
		// 上传图片并返回
		ele, err := uploadImage(c, groupCode, f, logger)
		if err != nil {
			logger.WithError(err).Error("Unable to upload image.")
			return nil, false, "网络错误，无法上传图片"
		}
		return ele, true, ""
	}

	logger.Debugf("No room for groupCode %d.", groupCode)
	return nil, false, "对局不存在"
}

func getChessString(room chessRoom) string {
	game := room.chessGame
	siteString := "[Site \"github.com/aimerneige/MiraiChess\"]\n"
	dataString := fmt.Sprintf("[Date \"%s\"]\n", time.Now().Format("2006-01-02"))
	whiteString := fmt.Sprintf("[White \"%s\"]\n", room.whiteName)
	blackString := fmt.Sprintf("[Black \"%s\"]\n", room.blackName)
	chessString := game.String()

	return siteString + dataString + whiteString + blackString + chessString
}

func getELOString(room chessRoom, whiteScore, blackScore float64, dbService *DBService) (string, error) {
	if room.whitePlayer == 0 || room.blackPlayer == 0 {
		return "", nil
	}
	eloString := "玩家等级分：\n"
	if err := updateELORate(room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName, whiteScore, blackScore, dbService); err != nil {
		eloString += "发生错误，无法更新等级分。"
		return eloString, err
	}
	whiteRate, blackRate, err := getELORate(room.whitePlayer, room.blackPlayer, dbService)
	if err != nil {
		eloString += "发生错误，无法获取等级分。"
		return eloString, err
	}
	eloString += fmt.Sprintf("%s：%d\n%s：%d\n\n", room.whiteName, whiteRate, room.blackName, blackRate)
	return eloString, nil
}

func getRankingString(dbService *DBService) (string, error) {
	eloList, err := dbService.GetHighestRateList()
	if err != nil {
		return "", err
	}
	ret := "当前等级分排行榜：\n\n"
	for _, elo := range eloList {
		ret += fmt.Sprintf("%s: %d\n", elo.Name, elo.Rate)
	}
	return ret, nil
}

// updateELORate 更新 elo 等级分
// 当数据库中没有玩家的等级分信息时，自动新建一条记录
func updateELORate(whiteUin, blackUin int64, whiteName, blackName string, whiteScore, blackScore float64, dbService *DBService) error {
	whiteRate, err := dbService.GetELORateByUin(whiteUin)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		// create white elo
		if err := dbService.CreateELO(whiteUin, whiteName, eloDefault); err != nil {
			return err
		}
		whiteRate = eloDefault
	}
	blackRate, err := dbService.GetELORateByUin(blackUin)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		// create black elo
		if err := dbService.CreateELO(blackUin, blackName, eloDefault); err != nil {
			return err
		}
		blackRate = eloDefault
	}
	whiteRate, blackRate = CalculateNewRate(whiteRate, blackRate, whiteScore, blackScore)
	// 更新白棋玩家的 ELO 等级分
	if err := dbService.UpdateELOByUin(whiteUin, whiteName, whiteRate); err != nil {
		return err
	}
	// 更新黑棋玩家的 ELO 等级分
	if err := dbService.UpdateELOByUin(blackUin, blackName, blackRate); err != nil {
		return err
	}

	return nil
}

// getELORate 获取玩家的 ELO 等级分
func getELORate(whiteUin, blackUin int64, dbService *DBService) (whiteRate int, blackRate int, err error) {
	whiteRate, err = dbService.GetELORateByUin(whiteUin)
	if err != nil {
		return
	}
	blackRate, err = dbService.GetELORateByUin(blackUin)
	if err != nil {
		return
	}
	return
}

func isAprilFoolsDay() bool {
	now := time.Now()
	return now.Month() == 4 && now.Day() == 1
}
