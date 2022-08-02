package service

import (
	"fmt"
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

const inkscapePath string = "./bin/inkscape"
const tempFileDir string = "./temp/"
const cheeseFilePath string = "./img/cheese.jpeg"
const board2svgScriptPath string = "./scripts/board2svg.py"

var instance *chessService
var eloEnabled bool = false
var eloDefault uint = 500

type chessService struct {
	gameRooms map[int64]chessRoom
}

type chessRoom struct {
	chessGame   *chess.Game
	whitePlayer int64
	whiteName   string
	blackPlayer int64
	blackName   string
	drawPlayer  int64
}

func init() {
	instance = &chessService{
		gameRooms: make(map[int64]chessRoom, 1),
	}
}

// Game 下棋
func Game(c *client.QQClient, groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		if room.blackPlayer != 0 {
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
		chessGame:   chess.NewGame(),
		whitePlayer: sender.Uin,
		whiteName:   sender.Nickname,
		blackPlayer: 0,
		blackName:   "",
		drawPlayer:  0,
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
	if groupMemberInfo.Permission != client.Administrator && groupMemberInfo.Permission != client.Owner {
		return nil
	}
	if room, ok := instance.gameRooms[groupCode]; ok {
		room.chessGame.Draw(chess.DrawOffer)
		chessString := getChessString(room)
		dbService := NewDBService(database.GetDB())
		if err := dbService.CreatePGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
			logger.WithError(err).Error("Fail to create PGN.")
		}
		delete(instance.gameRooms, groupCode)
		msg := simpleText("对局已被管理员中断，游戏结束。")
		if room.whitePlayer != 0 {
			msg.Append(message.NewAt(room.whitePlayer))
		}
		if room.blackPlayer != 0 {
			msg.Append(message.NewAt(room.blackPlayer))
		}
		msg.Append(message.NewText("\n\n" + chessString))
		return msg
	}
	return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
}

// Draw 和棋
func Draw(groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		if sender.Uin == room.whitePlayer || sender.Uin == room.blackPlayer {
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
			dbService := NewDBService(database.GetDB())
			if err := dbService.CreatePGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
				logger.WithError(err).Error("Fail to create PGN.")
			}
			whiteScore, blackScore := 0.5, 0.5
			eloString, err := getELOString(room, whiteScore, blackScore, dbService)
			if err != nil {
				logger.WithError(err).Error("Fail to get eloString. " + eloString)
			}
			delete(instance.gameRooms, groupCode)
			return textWithAt(sender.Uin, "接受和棋，游戏结束。\n"+eloString+chessString)
		}
		return textWithAt(sender.Uin, "不是对局中的玩家，无法请求和棋。")
	}
	return simpleText("对局不存在，发送「下棋」或「chess」可创建对局。")
}

// Resign 认输
func Resign(groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		// 检查是否是当前游戏玩家
		if sender.Uin == room.whitePlayer || sender.Uin == room.blackPlayer {
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
			eloString, err := getELOString(room, whiteScore, blackScore, dbService)
			if err != nil {
				logger.WithError(err).Error("Fail to get eloString. " + eloString)
			}
			delete(instance.gameRooms, groupCode)
			if isAprilFoolsDay() {
				return textWithAt(sender.Uin, "对手认输，游戏结束，你胜利了。\n"+eloString+chessString)
			}
			return textWithAt(sender.Uin, "认输，游戏结束。\n"+eloString+chessString)
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
			dbService := NewDBService(database.GetDB())
			if err := dbService.CreatePGN(chessString, room.whitePlayer, room.blackPlayer, room.whiteName, room.blackName); err != nil {
				logger.WithError(err).Error("Fail to create PGN.")
			}
			eloString, err := getELOString(room, whiteScore, blackScore, dbService)
			if err != nil {
				logger.WithError(err).Error("Fail to get eloString. " + eloString)
			}

			delete(instance.gameRooms, groupCode)
			return simpleText(msg + eloString + chessString).Append(boardImgEle)
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
	// 读取图片
	f, err := os.Open(cheeseFilePath)
	if err != nil {
		logger.WithError(err).Errorf("Unable to read open image file in %s.", cheeseFilePath)
		return nil
	}
	defer f.Close()
	// 上传图片
	ele, err := c.UploadGroupImage(groupCode, f)
	if err != nil {
		logger.WithError(err).Error("Unable to upload image.")
		return nil
	}
	return simpleText("Chess Cheese Cheese Chess").Append(ele)
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
		if err := exec.Command(board2svgScriptPath, room.chessGame.FEN(), svgFilePath, uciStr).Run(); err != nil {
			logger.Info(board2svgScriptPath, " ", room.chessGame.FEN(), " ", svgFilePath, " ", uciStr)
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
			logger.WithError(err).Errorf("Unable to read open image file in %s.", pngFilePath)
			return nil, false, "无法读取 png 图片"
		}
		defer f.Close()
		// 上传图片并返回
		ele, err := c.UploadGroupImage(groupCode, f)
		// 发生错误时重试 3 次，否则报错
		for i := 0; i < 3 && err != nil; i++ {
			ele, err = c.UploadGroupImage(groupCode, f)
		}
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
	eloString += fmt.Sprintf("%s：%d\n%s：%d\n", room.whiteName, whiteRate, room.blackName, blackRate)
	return eloString, nil
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
func getELORate(whiteUin, blackUin int64, dbService *DBService) (whiteRate uint, blackRate uint, err error) {
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
