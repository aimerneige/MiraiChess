package service

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/notnil/chess"
	"github.com/sirupsen/logrus"
)

const svgFilePath = "./temp/board.svg"
const pngFilePath = "./temp/board.png"

var instance *chessService

type chessService struct {
	gameRooms map[int64]chessRoom
}

type chessRoom struct {
	chessGame   *chess.Game
	whitePlayer int64
	blackPlayer int64
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
			return textWithAt(sender.Uin, "对局已在进行中，无法创建或加入对局。")
		}
		if sender.Uin == room.whitePlayer {
			return textWithAt(sender.Uin, "请等候其他玩家加入游戏。")
		}
		room.blackPlayer = sender.Uin
		instance.gameRooms[groupCode] = room
		boardImgEle, ok := getBoardElement(c, groupCode, logger)
		if !ok {
			delete(instance.gameRooms, groupCode)
			return errorResetText("无法生成棋盘图片。")
		}
		return simpleText("黑棋已加入对局，请白方下棋。").Append(message.NewAt(room.whitePlayer)).Append(boardImgEle)
	}
	instance.gameRooms[groupCode] = chessRoom{
		chessGame:   chess.NewGame(),
		whitePlayer: sender.Uin,
		blackPlayer: 0,
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
				return textWithAt(sender.Uin, "请求和棋，发送“和棋”或“draw”接受和棋。")
			}
			if room.drawPlayer == sender.Uin {
				return textWithAt(sender.Uin, "已发起和棋请求，请勿重复发送。")
			}
			delete(instance.gameRooms, groupCode)
			return textWithAt(sender.Uin, "接受和棋，游戏结束。"+room.chessGame.String())
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
			delete(instance.gameRooms, groupCode)
			return textWithAt(sender.Uin, "认输，游戏结束。"+room.chessGame.String())
		}
		return textWithAt(sender.Uin, "不是对局中的玩家，无法认输。")
	}
	return simpleText("对局不存在，发送“下棋”或“chess”可创建对局。")
}

// Play 走棋
func Play(c *client.QQClient, groupCode int64, sender *message.Sender, moveStr string, logger logrus.FieldLogger) *message.SendingMessage {
	if room, ok := instance.gameRooms[groupCode]; ok {
		// 不是对局中的玩家，忽略消息
		if (sender.Uin != room.whitePlayer) && (sender.Uin != room.blackPlayer) {
			// return textWithAt(sender.Uin, "您不是对局中的玩家，无法走棋。")
			return nil
		}
		// 对局未建立
		if room.blackPlayer == 0 {
			return textWithAt(room.whitePlayer, "请等候其他玩家加入游戏。")
		}
		// 需要对手走棋
		if ((sender.Uin == room.whitePlayer) && (room.chessGame.Position().Turn() != chess.White)) || ((sender.Uin == room.blackPlayer) && (room.chessGame.Position().Turn() != chess.Black)) {
			return textWithAt(sender.Uin, "请等待对手走棋。")
		}
		// 走棋
		if err := room.chessGame.MoveStr(moveStr); err != nil {
			return simpleText(fmt.Sprintf("移动“%s”违规，请检查，格式请参考“代数记谱法”(Algebraic notation)", moveStr))
		}
		// 走子之后，视为拒绝和棋
		if room.drawPlayer != 0 {
			room.drawPlayer = 0
			instance.gameRooms[groupCode] = room
		}
		// 生成棋盘图片
		boardImgEle, ok := getBoardElement(c, groupCode, logger)
		if !ok {
			delete(instance.gameRooms, groupCode)
			return errorResetText("无法生成棋盘图片")
		}
		if room.chessGame.Method() == chess.Stalemate {
			return simpleText("游戏结束，逼和。" + room.chessGame.String()).Append(boardImgEle)
		}
		if room.chessGame.Method() == chess.Checkmate {
			return simpleText("游戏结束，将杀。" + room.chessGame.String()).Append(boardImgEle)
		}
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

func errorResetText(errMsg string) *message.SendingMessage {
	return simpleText("发生错误，对局已重置。请联系开发者修 bug。\n开源地址 https://github.com/aimerneige/MiraiChess/issues\n错误信息：" + errMsg)
}

func simpleText(msg string) *message.SendingMessage {
	return message.NewSendingMessage().Append(message.NewText(msg))
}

func textWithAt(target int64, msg string) *message.SendingMessage {
	return message.NewSendingMessage().Append(message.NewAt(target)).Append(message.NewText(msg))
}

func getBoardElement(c *client.QQClient, groupCode int64, logger logrus.FieldLogger) (*message.GroupImageElement, bool) {
	if room, ok := instance.gameRooms[groupCode]; ok {
		if err := exec.Command("./module/chess/service/board_svg.py", room.chessGame.FEN(), svgFilePath).Run(); err != nil {
			logger.WithError(err).Error("Unable to generate svg file.")
			return nil, false
		}
		if err := exec.Command("./bin/inkscape", "-w", "720", "-h", "720", svgFilePath, "-o", pngFilePath).Run(); err != nil {
			logger.WithError(err).Error("Unable to convert to png.")
			return nil, false
		}
		f, err := os.Open(pngFilePath)
		if err != nil {
			logger.WithError(err).Errorf("Unable to read open image file in %s.", pngFilePath)
			return nil, false
		}
		defer f.Close()
		ele, err := c.UploadGroupImage(groupCode, f)
		if err != nil {
			logger.WithError(err).Error("Unable to upload image.")
			return nil, false
		}
		return ele, true
	}

	logger.Debugf("No room for groupCode %d.", groupCode)
	return nil, false
}
