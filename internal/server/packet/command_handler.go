package packet

import (
	logger "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server/database"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/utils"
	"strings"
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

	if user.Rating == int(Ban) {
		return ResultError(UserBaned, true, callsign)
	}

	if !user.VerifyPassword(password) {
		return ResultError(AuthFail, true, callsign)
	}
	c.User = user
	return nil
}

func (c *ConnectionHandler) handleAddAtc(data []string, rawLine []byte) *Result {
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

func (c *ConnectionHandler) handleAddPilot(data []string, rawLine []byte) *Result {
	//	#AP CES2352 SERVER 2352 123456  1   9  16  Half_nothing ZGHA
	//  [0] [  1  ] [  2 ] [ 3] [  4 ] [5] [6] [7] [       8       ]
	callsign := data[0]
	cid := getUserId(data[2])
	password := data[3]
	protocol := utils.StrToInt(data[5], 0)
	result := c.verifyUserInfo(callsign, protocol, cid, password)
	if result != nil {
		return result
	}
	reqRating := Rating(utils.StrToInt(data[4], 0))
	if reqRating != Observer {
		return ResultError(RequestLevelTooHigh, true, callsign)
	}
	simType := utils.StrToInt(data[6], 0)
	realName := data[7]
	if c.Client == nil {
		c.Client = NewClient(callsign, reqRating, cid, protocol, realName, c.Conn, false)
		c.Client.SimType = simType
		_ = clientManager.AddClient(c.Client)
	}
	clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	c.Client.SendMotd()
	logger.DebugF("[%s] Client login successfully", callsign)
	return ResultSuccess()
}

func (c *ConnectionHandler) handleAtcPosUpdate(data []string, rawLine []byte) *Result {
	//  %  ZSHA_CTR 24550  6  600  5  27.28025 118.28701  0
	// [0] [   1  ] [ 2 ] [3] [4] [5] [   6  ] [   7   ] [8]
	callsign := data[0]
	rating := Rating(utils.StrToInt(data[2], 0))
	facility := Facility(utils.StrToInt(data[4], 0))
	if !rating.CheckRatingFacility(facility) {
		return ResultError(RequestLevelTooHigh, true, callsign)
	}
	frequency := utils.StrToInt(data[1], 0)
	visualRange := utils.StrToFloat(data[3], 0)
	latitude := utils.StrToFloat(data[5], 0)
	longitude := utils.StrToFloat(data[6], 0)
	c.Client.UpdateAtcPos(frequency, facility, visualRange, latitude, longitude)
	clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	return ResultSuccess()
}

func (c *ConnectionHandler) handlePilotPosUpdate(data []string, rawLine []byte) *Result {
	//	@   S  CPA421 7000  1  38.96244 121.53479 87   0  4290770974 278
	// [0] [1] [  2 ] [ 3] [4] [   5  ] [   6   ] [7] [8] [    9   ] [10]
	transponder := utils.StrToInt(data[2], 0)
	latitude := utils.StrToFloat(data[4], 0)
	longitude := utils.StrToFloat(data[5], 0)
	altitude := utils.StrToInt(data[6], 0)
	groundSpeed := utils.StrToInt(data[7], 0)
	c.Client.UpdatePilotPos(transponder, latitude, longitude, altitude, groundSpeed)
	clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	return ResultSuccess()
}

func (c *ConnectionHandler) handleAtcVisPointUpdate(data []string, rawLine []byte) *Result {
	//  '  ZSHA_CTR  0  36.67349 120.45621
	// [0] [   1  ] [2] [   3  ] [   4   ]
	visPos := utils.StrToInt(data[1], 0)
	latitude := utils.StrToFloat(data[2], 0)
	longitude := utils.StrToFloat(data[3], 0)
	_ = c.Client.UpdateAtcVisPoint(visPos, latitude, longitude)
	clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	return ResultSuccess()
}

func (c *ConnectionHandler) sendFrequencyMessage(targetStation string, rawLine []byte) *Result {
	frequency := utils.StrToInt(targetStation[1:], -1)
	if frequency == -1 {
		return ResultError(Syntax, false, c.Client.Callsign)
	}
	if frequencyValid(frequency) {
		// 合法频率, 发给所有客户端
		clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	} else {
		// 非法频率, 大概率是管制使用, 发给管制
		clientManager.BroadcastMessage(rawLine, c.Client, CombineBroadcastFilter(BroadcastToAtc, BroadcastToClientInRange))
	}
	return nil
}

func (c *ConnectionHandler) handleClientQuery(data []string, rawLine []byte) *Result {
	//	查询飞行计划
	//	$CQ ZYSH_CTR SERVER FP  CPA421
	//  [0] [  1   ] [  2 ] [3] [  4 ]
	//
	//	修改飞行计划
	//	$CQ ZYSH_CTR @94835 FA  CPA421 31100
	//	[0] [  1   ] [  2 ] [3] [  4 ] [ 5 ]
	targetStation := data[1]
	if targetStation == "SERVER" {
		subQuery := data[2]
		if subQuery == "FP" {
			targetCallsign := data[3]
			client, err := clientManager.GetClient(targetCallsign)
			if err != nil {
				return ResultError(NoFlightPlan, false, c.Client.Callsign)
			}
			c.SendLine([]byte(client.FlightPlan.ToString(data[0])))
		}
	}
	if strings.HasPrefix(targetStation, "@") {
		result := c.sendFrequencyMessage(targetStation, rawLine)
		if result != nil {
			return result
		}
	} else {
		_ = clientManager.SendMessageTo(targetStation, rawLine)
	}
	return ResultSuccess()
}

func (c *ConnectionHandler) handleClientResponse(data []string, rawLine []byte) *Result {
	//	$CR ZSHA_CTR ZSSS_APP CAPS ATCINFO=1 SECPOS=1 MODELDESC=1 ONGOINGCOORD=1 NEWINFO=1 TEAMSPEAK=1 ICAOEQ=1
	//  [0] [   1  ] [   2  ] [ 3] [   4   ] [  5   ] [    6    ] [     7      ] [   8   ] [     9   ] [  10  ]
	//	$CR ZSHA_CTR SERVER ATIS  T  ZSHA_CTR Shanghai Control
	//	[0] [   1  ] [  2 ] [ 3] [4] [           5           ]
	targetStation := data[1]
	if targetStation == "SERVER" {
		subQuery := data[2]
		if subQuery == "ATIS" && data[3] == "T" {
			c.Client.AddAtcAtisInfo(data[4])
		}
	}
	if strings.HasPrefix(targetStation, "@") {
		result := c.sendFrequencyMessage(targetStation, rawLine)
		if result != nil {
			return result
		}
	} else {
		_ = clientManager.SendMessageTo(targetStation, rawLine)
	}
	return ResultSuccess()
}

func (c *ConnectionHandler) handleMessage(data []string, rawLine []byte) *Result {
	// #TM ZSHA_CTR ZSSS_APP 111
	// [0] [   1  ] [   2  ] [3]
	targetStation := data[1]
	if strings.HasPrefix(targetStation, "@") {
		result := c.sendFrequencyMessage(targetStation, rawLine)
		if result != nil {
			return result
		}
	} else {
		_ = clientManager.SendMessageTo(targetStation, rawLine)
	}
	return ResultSuccess()
}

func (c *ConnectionHandler) handlePlan(data []string, rawLine []byte) *Result {
	// $FP CPA421 SERVER  I  H/A320/L 474 ZYTL 1115  0  FL371 ZYHB  1    18   2    26  ZYCC
	// [0] [  1 ] [  2 ] [3] [  4   ] [5] [ 6] [ 7] [8] [ 9 ] [10] [11] [12] [13] [14] [15]
	// /V/ SEL/AHFL VENOS A588 NULRA W206 MAGBI W656 ISLUK W629 LARUN
	// [    16    ] [                      17                       ]
	err := c.Client.UpdateFlightPlan(data)
	if err != nil {
		return ResultError(Syntax, false, c.Client.Callsign)
	}
	if !c.Client.FlightPlan.Locked {
		clientManager.BroadcastMessage(rawLine, c.Client, CombineBroadcastFilter(BroadcastToAtc, BroadcastToClientInRange))
	}
	return ResultSuccess()
}

func (c *ConnectionHandler) handleAtcEditPlan(data []string, rawLine []byte) *Result {
	// $AM ZYSH_CTR SERVER CPA421  I  H/A320/L 474 ZYTL 1115  0  FL371 ZYHB  11  8     22   6   ZYCC
	// [0] [   1  ] [  2 ] [  3 ] [4] [   5  ] [6] [ 7] [ 8] [9] [ 10] [11] [12] [13] [14] [15] [16]
	// /V/ SEL/AHFL CHI19D/28 VENOS A588 NULRA W206 MAGBI W656 ISLUK W629 LARUN
	// [     17   ] [                             18                          ]
	if !c.Client.IsAtc {
		return ResultError(Syntax, false, c.Client.Callsign)
	}
	if !allowEditPlanFacility.CheckFacility(c.Client.Facility) {
		return ResultError(Syntax, false, c.Client.Callsign)
	}
	targetCallsign := data[2]
	client, err := clientManager.GetClient(targetCallsign)
	if err != nil {
		return ResultError(Syntax, false, c.Client.Callsign)
	}
	if client.FlightPlan == nil {
		return ResultError(NoFlightPlan, false, c.Client.Callsign)
	}
	client.FlightPlan.Locked = true
	err = client.UpdateFlightPlan(data[1:])
	if err != nil {
		return ResultError(Syntax, false, c.Client.Callsign)
	}
	clientManager.BroadcastMessage([]byte(client.FlightPlan.ToString(string(AllATC))),
		c.Client, CombineBroadcastFilter(BroadcastToAtc, BroadcastToClientInRange))
	return ResultSuccess()
}

//func (c *ConnectionHandler) handleKillClient(data []string, rawLine []byte) *Result {
//
//}

func (c *ConnectionHandler) handleRequest(data []string, rawLine []byte) *Result {
	// $HO ZSSS_APP ZSHA_CTR CES2352
	// [0] [  1   ] [   2  ] [  3  ]
	// #PC ZSHA_CTR ZSSS_APP CCP HC CES2352
	// $HA ZSHA_CTR ZSSS_APP CES2352
	targetStation := data[1]
	_ = clientManager.SendMessageTo(targetStation, rawLine)
	return ResultSuccess()
}

func (c *ConnectionHandler) HandleCommand(commandType ClientCommand, data []string, rawLine []byte) *Result {
	var result = ResultSuccess()
	switch commandType {
	case AddAtc:
		result = c.handleAddAtc(data, rawLine)
		break
	case RemoveAtc:
		break
	case AddPilot:
		result = c.handleAddPilot(data, rawLine)
		break
	case RemovePilot:
		break
	case RequestHandoff:
	case AcceptHandoff:
	case ProController:
		result = c.handleRequest(data, rawLine)
		break
	case PilotPosition:
		result = c.handlePilotPosUpdate(data, rawLine)
		break
	case Plan:
		result = c.handlePlan(data, rawLine)
		break
	case AtcEditPlan:
		result = c.handleAtcEditPlan(data, rawLine)
		break
	case KillClient:
		break
	case AtcPosition:
		result = c.handleAtcPosUpdate(data, rawLine)
		break
	case AtcSubVisPoint:
		result = c.handleAtcVisPointUpdate(data, rawLine)
		break
	case Message:
		result = c.handleMessage(data, rawLine)
		break
	case ClientQuery:
		result = c.handleClientQuery(data, rawLine)
		break
	case ClientResponse:
		result = c.handleClientResponse(data, rawLine)
		break
	default:
		result = ResultSuccess()
	}
	return result
}
