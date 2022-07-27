package service

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/notnil/chess"
	"github.com/sirupsen/logrus"
)

var inkscapePath string = "./bin/inkscape"
var svgFilePath string = "./temp/board.svg"
var pngFilePath string = "./temp/board.png"
var cheeseFilePath string = "./img/cheese.jpeg"
var board2svgScriptPath string = "./scripts/board2svg.py"

var instance *chessService

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

// UpdatePathConfig update path config
func UpdatePathConfig(inkscape, svg, png, cheese, script string) {
	inkscapePath = inkscapePath
	svgFilePath = svg
	pngFilePath = png
	board2svgScriptPath = script
}

// Game 下棋
func Game(c *client.QQClient, groupCode int64, sender *message.Sender, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		if room.blackPlayer != 0 {
			return textWithAt(sender.Uin, "对局已在进行中，无法创建或加入对局。")
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
	return simpleText("已创建新的对局，发送“下棋”或“chess”可加入对局。")
}

// Draw 和棋
func Draw(groupCode int64, sender *message.Sender) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		if sender.Uin == room.whitePlayer || sender.Uin == room.blackPlayer {
			if room.drawPlayer == 0 {
				room.drawPlayer = sender.Uin
				instance.gameRooms[groupCode] = room
				return textWithAt(sender.Uin, "请求和棋，发送“和棋”或“draw”接受和棋。走棋视为拒绝和棋。")
			}
			if room.drawPlayer == sender.Uin {
				return textWithAt(sender.Uin, "已发起和棋请求，请勿重复发送。")
			}
			room.chessGame.Draw(chess.DrawOffer)
			chessString := getChessString(room)
			delete(instance.gameRooms, groupCode)
			return textWithAt(sender.Uin, "接受和棋，游戏结束。\n"+chessString)
		}
		return textWithAt(sender.Uin, "不是对局中的玩家，无法请求和棋。")
	}
	return simpleText("对局不存在，发送“下棋”或“chess”可创建对局。")
}

// Resign 认输
func Resign(groupCode int64, sender *message.Sender) *message.SendingMessage {
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
			delete(instance.gameRooms, groupCode)
			if isAprilFoolsDay() {
				return textWithAt(sender.Uin, "对手认输，游戏结束，你胜利了。\n"+chessString)
			}
			return textWithAt(sender.Uin, "认输，游戏结束。\n"+chessString)
		}
		return textWithAt(sender.Uin, "不是对局中的玩家，无法认输。")
	}
	return simpleText("对局不存在，发送“下棋”或“chess”可创建对局。")
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
			return simpleText(fmt.Sprintf("移动“%s”违规，请检查，格式请参考“代数记谱法”(Algebraic notation)。", moveStr))
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
					winner = "黑方"
				} else {
					winner = "白方"
				}
				msg += winner
				msg += "胜利，因为将杀。\n"
			}
			chessString := getChessString(room)
			delete(instance.gameRooms, groupCode)
			return simpleText(msg + chessString).Append(boardImgEle)
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
	return textWithAt(sender.Uin, "对局不存在，发送“下棋”或“chess”可创建对局。")
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
	return simpleText("发生错误，请联系开发者修 bug。\n开源地址 https://github.com/aimerneige/MiraiChess/issues\n错误信息：" + errMsg)
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
	dataString := fmt.Sprintf("[Date \"%s\"]\n", time.Now().Format("2006-01-02"))
	whiteString := fmt.Sprintf("[White \"%s\"]\n", room.whiteName)
	blackString := fmt.Sprintf("[Black \"%s\"]\n", room.blackName)
	chessString := game.String()

	return dataString + whiteString + blackString + chessString
}

func isAprilFoolsDay() bool {
	now := time.Now()
	return now.Month() == 4 && now.Day() == 1
}
