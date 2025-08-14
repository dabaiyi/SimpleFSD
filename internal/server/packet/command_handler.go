// Package packet 命令处理的核心函数定义
package packet

import (
	"fmt"
	logger "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/utils"
	"strings"
)

// getUserId 转换客户端传递过来的cid为数据库可识别的模式
func getUserId(cid string) database.UserId {
	id := utils.StrToInt(cid, -1)
	if id != -1 {
		return database.IntUserId(id)
	}
	return database.StringUserId(cid)
}

func (c *ConnectionHandler) checkPacketLength(data []string, requirement *CommandRequirement) (*Result, bool) {
	length := len(data)
	if length < requirement.RequireLength {
		return resultError(Syntax, requirement.Fatal, c.Callsign, fmt.Errorf("datapack length too short, require %d but got %d", requirement.RequireLength, length)), false
	}
	return nil, true
}

// verifyUserInfo 验证用户信息与处理客户端重连机制
func (c *ConnectionHandler) verifyUserInfo(callsign string, protocol int, cid database.UserId, password string) *Result {
	if !callsignValid(callsign) {
		return resultError(CallsignInvalid, true, callsign, nil)
	}

	if protocol != 9 {
		return resultError(InvalidProtocolVision, true, callsign, nil)
	}

	client, ok := clientManager.GetClient(callsign)

	// 客户端存在且标记为断开连接
	if ok {
		if client.Reconnect(c) {
			// 客户端重连
			c.Client = client
		} else {
			// 呼号已被使用
			return resultError(CallsignInUse, true, callsign, nil)
		}
	}

	user, err := cid.GetUser()
	if err != nil {
		return resultError(AuthFail, true, callsign, err)
	}
	if user.Rating == Ban.Index() {
		return resultError(UserBaned, true, callsign, nil)
	}
	if !user.VerifyPassword(password) {
		return resultError(AuthFail, true, callsign, nil)
	}

	// 重设重连客户端的User
	if client != nil {
		client.User = user
	}
	c.User = user

	return nil
}

// handleAddAtc 处理管制员登录
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
		return resultError(RequestLevelTooHigh, true, callsign, nil)
	}
	realName := data[2]
	latitude := utils.StrToFloat(data[9], 0)
	longitude := utils.StrToFloat(data[10], 0)
	if c.Client == nil {
		c.Client = NewClient(callsign, Rating(reqRating), protocol, realName, c, true)
		_ = c.Client.SetPosition(0, latitude, longitude)
		_ = clientManager.AddClient(c.Client)
	}
	c.Client.SendLine(makePacket(ClientQuery, "SERVER", callsign, "ATIS"))
	go clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	c.Client.SendMotd()
	logger.InfoF("[%s] ATC login successfully", callsign)
	return resultSuccess()
}

// handleAddPilot 处理客户端登录
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
	reqRating := Rating(utils.StrToInt(data[4], 0) - 1)
	if reqRating != Normal || !ratingFacilityMap[reqRating].CheckFacility(Pilot) {
		return resultError(RequestLevelTooHigh, true, callsign, nil)
	}
	simType := utils.StrToInt(data[6], 0)
	realName := data[7]
	if c.Client == nil {
		c.Client = NewClient(callsign, reqRating, protocol, realName, c, false)
		c.Client.SimType = simType
		_ = clientManager.AddClient(c.Client)
	}
	go clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	c.Client.SendMotd()
	logger.InfoF("[%s] Client login successfully", callsign)
	if !config.Server.General.SimulatorServer && c.Client.FlightPlan != nil &&
		c.Client.FlightPlan.FromWeb && callsign != c.Client.FlightPlan.Callsign {
		c.Client.SendLine(makePacket(Message, "FPlanManager", callsign,
			fmt.Sprintf("Seems you are connect with callsign(%s), "+
				"but we found a flightplan submit by web at %s which has callsign(%s), "+
				"please check it.", c.Client.FlightPlan.UpdatedAt.String(), callsign, c.Client.FlightPlan.Callsign)))
	}
	return resultSuccess()
}

// handleAtcPosUpdate 处理管制员位置更新
func (c *ConnectionHandler) handleAtcPosUpdate(data []string, rawLine []byte) *Result {
	//  %  ZSHA_CTR 24550  6  600  5  27.28025 118.28701  0
	// [0] [   1  ] [ 2 ] [3] [4] [5] [   6  ] [   7   ] [8]
	callsign := data[0]
	facility := Facility(1 << utils.StrToInt(data[2], 0))
	rating := Rating(utils.StrToInt(data[4], 0))
	if !rating.CheckRatingFacility(facility) {
		return resultError(RequestLevelTooHigh, true, callsign, nil)
	}
	frequency := utils.StrToInt(data[1], 0)
	visualRange := utils.StrToFloat(data[3], 0)
	latitude := utils.StrToFloat(data[5], 0)
	longitude := utils.StrToFloat(data[6], 0)
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	go clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	c.Client.UpdateAtcPos(frequency, facility, visualRange, latitude, longitude)
	return resultSuccess()
}

// handlePilotPosUpdate 处理飞行员位置更新
func (c *ConnectionHandler) handlePilotPosUpdate(data []string, rawLine []byte) *Result {
	//	@   S  CPA421 7000  1  38.96244 121.53479 87   0  4290770974 278
	// [0] [1] [  2 ] [ 3] [4] [   5  ] [   6   ] [7] [8] [    9   ] [10]
	transponder := utils.StrToInt(data[2], 0)
	latitude := utils.StrToFloat(data[4], 0)
	longitude := utils.StrToFloat(data[5], 0)
	altitude := utils.StrToInt(data[6], 0)
	groundSpeed := utils.StrToInt(data[7], 0)
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	go clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	c.Client.UpdatePilotPos(transponder, latitude, longitude, altitude, groundSpeed)
	return resultSuccess()
}

// handleAtcVisPointUpdate 处理管制员视程点更新
func (c *ConnectionHandler) handleAtcVisPointUpdate(data []string, _ []byte) *Result {
	//  '  ZSHA_CTR  0  36.67349 120.45621
	// [0] [   1  ] [2] [   3  ] [   4   ]
	visPos := utils.StrToInt(data[1], 0)
	latitude := utils.StrToFloat(data[2], 0)
	longitude := utils.StrToFloat(data[3], 0)
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	_ = c.Client.UpdateAtcVisPoint(visPos, latitude, longitude)
	return resultSuccess()
}

// sendFrequencyMessage 发送频率消息
func (c *ConnectionHandler) sendFrequencyMessage(targetStation string, rawLine []byte) *Result {
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	frequency := utils.StrToInt(targetStation[1:], -1)
	if frequency == -1 {
		return resultError(Syntax, false, c.Client.Callsign, fmt.Errorf("illegal frequency %s", targetStation))
	}
	if frequencyValid(frequency) {
		// 合法频率, 发给所有客户端
		go clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToClientInRange)
	} else {
		// 非法频率, 大概率是管制使用, 只发给管制
		go clientManager.BroadcastMessage(rawLine, c.Client, CombineBroadcastFilter(BroadcastToAtc, BroadcastToClientInRange))
	}
	return nil
}

// handleClientQuery 处理客户端查询消息
func (c *ConnectionHandler) handleClientQuery(data []string, rawLine []byte) *Result {
	//	查询飞行计划
	//	$CQ ZYSH_CTR SERVER FP  CPA421
	//  [0] [  1   ] [  2 ] [3] [  4 ]
	//
	//	修改飞行计划
	//	$CQ ZYSH_CTR @94835 FA  CPA421 31100
	//	[0] [  1   ] [  2 ] [3] [  4 ] [ 5 ]
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	commandLength := len(data)
	targetStation := data[1]
	if targetStation == "SERVER" {
		subQuery := data[2]
		// 查询指定机组的飞行计划
		if subQuery == "FP" {
			targetCallsign := data[3]
			client, ok := clientManager.GetClient(targetCallsign)
			if !ok || client.FlightPlan == nil {
				return resultError(NoFlightPlan, false, c.Client.Callsign, nil)
			}
			c.Client.SendLine([]byte(client.FlightPlan.ToString(data[0])))
		}
	}
	// 如果发送目标是一个频率
	if strings.HasPrefix(targetStation, "@") {
		// 如果目标频率是94835
		if !config.Server.General.SimulatorServer && targetStation == specialFrequency {
			// 这里并不是发给服务器的, 所以如果客户端没有权限, 直接返回就行
			if !c.Client.CheckFacility(allowAtcFacility) {
				return resultSuccess()
			}
			subQuery := data[2]
			if subQuery == "FA" && commandLength >= 5 {
				targetCallsign := data[3]
				client, ok := clientManager.GetClient(targetCallsign)
				if !ok {
					// 这里并不是发给服务器的, 所以如果找不到指定客户端, 直接返回就行
					return resultSuccess()
				}
				cruiseAltitude := utils.StrToInt(data[4], 0)
				if err := client.FlightPlan.UpdateCruiseAltitude(fmt.Sprintf("FL%03d", cruiseAltitude/100), true); err != nil {
					// 这里并不是发给服务器的, 所以如果出错, 直接返回就行
					return resultSuccess()
				}
			}
		}
		err := c.sendFrequencyMessage(targetStation, rawLine)
		if err != nil {
			return err
		}
	} else {
		_ = clientManager.SendMessageTo(targetStation, rawLine)
	}
	return resultSuccess()
}

// handleClientResponse 处理客户端回复消息
func (c *ConnectionHandler) handleClientResponse(data []string, rawLine []byte) *Result {
	//	$CR ZSHA_CTR ZSSS_APP CAPS ATCINFO=1 SECPOS=1 MODELDESC=1 ONGOINGCOORD=1 NEWINFO=1 TEAMSPEAK=1 ICAOEQ=1
	//  [0] [   1  ] [   2  ] [ 3] [   4   ] [  5   ] [    6    ] [     7      ] [   8   ] [     9   ] [  10  ]
	//	$CR ZSHA_CTR SERVER ATIS  T  ZSHA_CTR Shanghai Control
	//	[0] [   1  ] [  2 ] [ 3] [4] [           5           ]
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	commandLength := len(data)
	targetStation := data[1]
	if targetStation == "SERVER" {
		subQuery := data[2]
		if subQuery == "ATIS" && commandLength >= 5 && data[3] == "T" {
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
	return resultSuccess()
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
	} else if strings.HasPrefix(targetStation, "*") {
		// 广播消息
		if targetStation == string(AllSup) {
			go clientManager.BroadcastMessage(rawLine, c.Client, BroadcastToSup)
		}
	} else {
		_ = clientManager.SendMessageTo(targetStation, rawLine)
	}
	return resultSuccess()
}

func (c *ConnectionHandler) handlePlan(data []string, rawLine []byte) *Result {
	// $FP CPA421 SERVER  I  H/A320/L 474 ZYTL 1115  0  FL371 ZYHB  1    18   2    26  ZYCC
	// [0] [  1 ] [  2 ] [3] [  4   ] [5] [ 6] [ 7] [8] [ 9 ] [10] [11] [12] [13] [14] [15]
	// /V/ SEL/AHFL VENOS A588 NULRA W206 MAGBI W656 ISLUK W629 LARUN
	// [    16    ] [                      17                       ]
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	if c.Client.IsAtc {
		return resultError(Syntax, false, c.Client.Callsign, fmt.Errorf("atc can not submit fligth plan"))
	}
	if err := c.Client.UpsertFlightPlan(data); err != nil {
		return resultError(Syntax, false, c.Client.Callsign, err)
	}
	if !c.Client.FlightPlan.Locked {
		go clientManager.BroadcastMessage(rawLine, c.Client, CombineBroadcastFilter(BroadcastToAtc, BroadcastToClientInRange))
	}
	return resultSuccess()
}

func (c *ConnectionHandler) handleAtcEditPlan(data []string, _ []byte) *Result {
	// $AM ZYSH_CTR SERVER CPA421  I  H/A320/L 474 ZYTL 1115  0  FL371 ZYHB  11  8     22   6   ZYCC
	// [0] [   1  ] [  2 ] [  3 ] [4] [   5  ] [6] [ 7] [ 8] [9] [ 10] [11] [12] [13] [14] [15] [16]
	// /V/ SEL/AHFL CHI19D/28 VENOS A588 NULRA W206 MAGBI W656 ISLUK W629 LARUN
	// [     17   ] [                             18                          ]
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	if !c.Client.IsAtc {
		return resultError(Syntax, false, c.Client.Callsign, fmt.Errorf("only act can edit flight plan"))
	}
	if !c.Client.CheckFacility(allowAtcFacility) {
		return resultError(Syntax, false, c.Client.Callsign, fmt.Errorf("%s facility not allowed to edit plan", c.Client.Facility.String()))
	}
	targetCallsign := data[2]
	client, ok := clientManager.GetClient(targetCallsign)
	if !ok {
		return resultError(SourceCallsignInvalid, false, c.Client.Callsign, fmt.Errorf("%s not exists", targetCallsign))
	}
	if client.FlightPlan == nil {
		return resultError(NoFlightPlan, false, c.Client.Callsign, fmt.Errorf("%s do not have filght plan", c.Client.Callsign))
	}
	client.FlightPlan.Locked = !config.Server.General.SimulatorServer
	if err := client.FlightPlan.UpdateFlightPlan(data[1:], true); err != nil {
		return resultError(Syntax, false, c.Client.Callsign, err)
	}
	go clientManager.BroadcastMessage([]byte(client.FlightPlan.ToString(string(AllATC))),
		c.Client, CombineBroadcastFilter(BroadcastToAtc, BroadcastToClientInRange))
	return resultSuccess()
}

func (c *ConnectionHandler) handleKillClient(data []string, _ []byte) *Result {
	// $!! ZSHA_CTR CPA421 test
	if c.Client == nil {
		return resultError(Syntax, false, "", fmt.Errorf("client not register"))
	}
	if !(c.Client.IsAtc && c.Client.CheckRating(allowKillRating)) {
		return resultError(Syntax, false, c.Client.Callsign, fmt.Errorf("%s facility not allowed to kill client", c.Client.Facility.String()))
	}
	targetStation := data[1]
	client, ok := clientManager.GetClient(targetStation)
	if !ok {
		return resultError(SourceCallsignInvalid, false, c.Client.Callsign, fmt.Errorf("%s not exists", targetStation))
	}
	client.MarkedDisconnect(false)
	return resultSuccess()
}

func (c *ConnectionHandler) handleRequest(data []string, rawLine []byte) *Result {
	// $HO ZSSS_APP ZSHA_CTR CES2352
	// [0] [  1   ] [   2  ] [  3  ]
	// #PC ZSHA_CTR ZSSS_APP CCP HC CES2352
	// $HA ZSHA_CTR ZSSS_APP CES2352
	targetStation := data[1]
	_ = clientManager.SendMessageTo(targetStation, rawLine)
	return resultSuccess()
}

func (c *ConnectionHandler) removeClient(_ []string, _ []byte) *Result {
	// #DA ZGGG_CTR SERVER
	logger.InfoF("[%s] Offline", c.Client.Callsign)
	return resultSuccess()
}

func (c *ConnectionHandler) handleCommand(commandType ClientCommand, data []string, rawLine []byte) *Result {
	var result = resultSuccess()
	if requirement, ok := CommandRequirements[commandType]; ok {
		if err, ok := c.checkPacketLength(data, requirement); !ok {
			return err
		}
	}
	switch commandType {
	case AddAtc:
		result = c.handleAddAtc(data, rawLine)
	case AddPilot:
		result = c.handleAddPilot(data, rawLine)
	case RequestHandoff, AcceptHandoff, ProController:
		result = c.handleRequest(data, rawLine)
	case PilotPosition:
		result = c.handlePilotPosUpdate(data, rawLine)
	case Plan:
		result = c.handlePlan(data, rawLine)
	case AtcEditPlan:
		result = c.handleAtcEditPlan(data, rawLine)
	case KillClient:
		result = c.handleKillClient(data, rawLine)
	case AtcPosition:
		result = c.handleAtcPosUpdate(data, rawLine)
	case AtcSubVisPoint:
		result = c.handleAtcVisPointUpdate(data, rawLine)
	case Message:
		result = c.handleMessage(data, rawLine)
	case ClientQuery:
		result = c.handleClientQuery(data, rawLine)
	case ClientResponse:
		result = c.handleClientResponse(data, rawLine)
	case RemoveAtc, RemovePilot:
		c.removeClient(data, rawLine)
	default:
		result = resultSuccess()
	}
	return result
}
