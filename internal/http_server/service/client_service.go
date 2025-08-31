// Package service
package service

import (
	"errors"
	c "github.com/half-nothing/simple-fsd/internal/config"
	"github.com/half-nothing/simple-fsd/internal/interfaces/fsd"
	"github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"github.com/half-nothing/simple-fsd/internal/utils"
	"time"
)

type ClientService struct {
	onlineClient  *utils.CachedValue[OnlineClients]
	clientManager fsd.ClientManagerInterface
	emailService  EmailServiceInterface
	config        *c.HttpServerConfig
	userOperation operation.UserOperationInterface
}

func NewClientService(
	config *c.HttpServerConfig,
	clientManager fsd.ClientManagerInterface,
	emailService EmailServiceInterface,
	userOperation operation.UserOperationInterface,
) *ClientService {
	service := &ClientService{
		clientManager: clientManager,
		emailService:  emailService,
		config:        config,
		userOperation: userOperation,
	}
	service.onlineClient = utils.NewCachedValue[OnlineClients](config.CacheDuration, func() *OnlineClients { return service.getOnlineClient() })
	return service
}

func (clientService *ClientService) getOnlineClient() *OnlineClients {
	data := &OnlineClients{
		General: OnlineGeneral{
			Version:          3,
			ConnectedClients: 0,
			OnlinePilot:      0,
			OnlineController: 0,
		},
		Pilots:      make([]*OnlinePilot, 0),
		Controllers: make([]*OnlineController, 0),
	}

	clientCopy := clientService.clientManager.GetClientSnapshot()
	defer clientService.clientManager.PutSlice(clientCopy)

	for _, client := range clientCopy {
		if client == nil || client.Disconnected() {
			continue
		}
		data.General.ConnectedClients++
		if client.IsAtc() {
			data.General.OnlineController++
			controller := &OnlineController{
				Cid:       client.User().Cid,
				Callsign:  client.Callsign(),
				RealName:  client.RealName(),
				Latitude:  client.Position()[0].Latitude,
				Longitude: client.Position()[0].Longitude,
				Rating:    client.Rating().Index(),
				Facility:  client.Facility().Index(),
				Frequency: client.Frequency() + 100000,
				AtcInfo:   client.AtisInfo(),
				LogonTime: client.History().StartTime.Format(time.DateTime),
			}
			data.Controllers = append(data.Controllers, controller)
		} else {
			data.General.OnlinePilot++
			pilot := &OnlinePilot{
				Cid:         client.User().Cid,
				Callsign:    client.Callsign(),
				RealName:    client.RealName(),
				Latitude:    client.Position()[0].Latitude,
				Longitude:   client.Position()[0].Longitude,
				Transponder: client.Transponder(),
				Heading:     client.Heading(),
				Altitude:    client.Altitude(),
				GroundSpeed: client.GroundSpeed(),
				FlightPlan:  client.FlightPlan(),
				LogonTime:   client.History().StartTime.Format(time.DateTime),
			}
			data.Pilots = append(data.Pilots, pilot)
		}
	}

	data.General.GenerateTime = time.Now().Format(time.DateTime)

	return data
}

func (clientService *ClientService) GetOnlineClient() *OnlineClients {
	return clientService.onlineClient.GetValue()
}

var (
	ErrCallsignNotFound = ApiStatus{StatusName: "CALLSIGN_NOT_FOUND", Description: "发送目标不在线", HttpCode: NotFound}
	SuccessSendMessage  = ApiStatus{StatusName: "SEND_MESSAGE", Description: "发送成功", HttpCode: Ok}
)

func (clientService *ClientService) SendMessageToClient(req *RequestSendMessageToClient) *ApiResponse[ResponseSendMessageToClient] {
	if req.Uid <= 0 || req.SendTo == "" || req.Message == "" {
		return NewApiResponse[ResponseSendMessageToClient](&ErrIllegalParam, Unsatisfied, nil)
	}
	if req.Permission <= 0 {
		return NewApiResponse[ResponseSendMessageToClient](&ErrNoPermission, Unsatisfied, nil)
	}
	permission := operation.Permission(req.Permission)
	if !permission.HasPermission(operation.ClientSendMessage) {
		return NewApiResponse[ResponseSendMessageToClient](&ErrNoPermission, Unsatisfied, nil)
	}
	if err := clientService.clientManager.SendRawMessageTo(req.Cid, req.SendTo, req.Message); errors.Is(err, fsd.ErrCallsignNotFound) {
		return NewApiResponse[ResponseSendMessageToClient](&ErrCallsignNotFound, Unsatisfied, nil)
	}
	data := ResponseSendMessageToClient(true)
	return NewApiResponse[ResponseSendMessageToClient](&SuccessSendMessage, Unsatisfied, &data)
}

var SuccessKillClient = ApiStatus{StatusName: "KILL_CLIENT", Description: "成功踢出客户端", HttpCode: Ok}

func (clientService *ClientService) KillClient(req *RequestKillClient) *ApiResponse[ResponseKillClient] {
	if req.Uid <= 0 || req.TargetCallsign == "" {
		return NewApiResponse[ResponseKillClient](&ErrIllegalParam, Unsatisfied, nil)
	}
	user, res := CallDBFuncAndCheckError[operation.User, ResponseKillClient](func() (*operation.User, error) {
		return clientService.userOperation.GetUserByUid(req.Uid)
	})
	if res != nil {
		return res
	}
	permission := operation.Permission(user.Permission)
	if !permission.HasPermission(operation.ClientKill) {
		return NewApiResponse[ResponseKillClient](&ErrNoPermission, Unsatisfied, nil)
	}
	client, ok := clientService.clientManager.GetClient(req.TargetCallsign)
	if !ok {
		return NewApiResponse[ResponseKillClient](&ErrCallsignNotFound, Unsatisfied, nil)
	}
	client.MarkedDisconnect(false)
	if clientService.config.Email.Template.EnableKickedFromServerEmail {
		if err := clientService.emailService.SendKickedFromServerEmail(client.User(), user, req.Reason); err != nil {
			c.ErrorF("SendRatingChangeEmail Failed: %v", err)
		}
	}
	data := ResponseKillClient(true)
	return NewApiResponse[ResponseKillClient](&SuccessKillClient, Unsatisfied, &data)
}

var (
	ErrClientNotFound    = ApiStatus{StatusName: "CLIENT_NOT_FOUND", Description: "指定客户端不存在", HttpCode: NotFound}
	SuccessGetClientPath = ApiStatus{StatusName: "GET_CLIENT_PATH", Description: "获取指定客户端飞行路径", HttpCode: Ok}
)

func (clientService *ClientService) GetClientPath(req *RequestClientPath) *ApiResponse[ResponseClientPath] {
	if req.Callsign == "" {
		return NewApiResponse[ResponseClientPath](&ErrIllegalParam, Unsatisfied, nil)
	}
	client, exist := clientService.clientManager.GetClient(req.Callsign)
	if !exist {
		return NewApiResponse[ResponseClientPath](&ErrClientNotFound, Unsatisfied, nil)
	}
	data := ResponseClientPath(client.Paths())
	return NewApiResponse(&SuccessGetClientPath, Unsatisfied, &data)
}
