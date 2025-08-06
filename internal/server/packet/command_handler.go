package packet

import (
	logger "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server/database"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/utils"
)

func getUserId(cid string) database.UserId {
	id := utils.StrToInt(cid, -1)
	if id != -1 {
		return database.IntId(id)
	}
	return database.StringId(cid)
}

func (c *ConnectionHandler) verifyUserInfo(callsign string, protocol int, cid database.UserId, password string) *Result {
	if !callsignValid(callsign) {
		return ResultError(CallsignInvalid, true, callsign)
	}
	if protocol != 9 {
		return ResultError(InvalidProtocolVision, true, callsign)
	}

	client, _ := clientManager.GetClient(callsign)

	// 客户端存在且标记为断开连接
	if client != nil {
		if client.Reconnect(c.Conn) {
			// 客户端重连
			c.Client = client
		} else {
			// 呼号已被使用
			return ResultError(CallsignInUse, true, callsign)
		}
	}

	user, err := cid.GetUser()
	if err != nil {
		return ResultError(AuthFail, true, callsign)
	}

	if !user.VerifyPassword(password) {
		return ResultError(AuthFail, true, callsign)
	}
	c.User = user
	return nil
}

func (c *ConnectionHandler) handleAddAtcCommand(data []string, rawLine []byte) *Result {
	// #AA 2352_OBS SERVER 2352 2352 123456  1  9  1  0  29.86379 119.49287 100
	// [0] [   1  ] [  2 ] [ 3] [ 4] [  5 ] [6][7][8][9] [  10  ] [   11  ] [12]
	callsign := data[0]
	cid := getUserId(data[3])
	password := data[4]
	protocol := utils.StrToInt(data[6], 0)
	result := c.verifyUserInfo(callsign, protocol, cid, password)
	if result != nil {
		return result
	}
	reqRating := utils.StrToInt(data[5], 0)
	if reqRating > c.User.Rating {
		return ResultError(RequestLevelTooHigh, true, callsign)
	}
	realName := data[2]
	latitude := utils.StrToFloat(data[9], 0)
	longitude := utils.StrToFloat(data[10], 0)
	if c.Client == nil {
		c.Client = NewClient(callsign, Rating(reqRating), cid, protocol, realName, c.Conn, true)
		_ = c.Client.SetPosition(0, latitude, longitude)
		_ = clientManager.AddClient(c.Client)
	}
	_, _ = c.Conn.Write(makePacket(ClientQuery, "SERVER", callsign, "ATIS"))
	clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	c.Client.SendMotd()
	logger.DebugF("[%s] Client login successfully", callsign)
	return ResultSuccess()
}

func (c *ConnectionHandler) HandleCommand(commandType ClientCommand, data []string, rawLine []byte) *Result {
	var result = ResultSuccess()
	switch commandType {
	case AddAtc:
		result = c.handleAddAtcCommand(data, rawLine)
		break
	case RemoveAtc:
		break
	case AddPilot:
		break
	case RemovePilot:
		break
	case RequestHandoff:
		break
	case AcceptHandoff:
		break
	case ProController:
		break
	case PilotPosition:
		break
	case AtcPosition:
		break
	case AtcSubVisPoint:
		break
	case Message:
		break
	case ClientQuery:
		break
	case ClientResponse:
		break
	default:
		result = ResultSuccess()
	}
	return result
}
