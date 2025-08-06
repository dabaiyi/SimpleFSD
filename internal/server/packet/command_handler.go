package packet

func handleAddAtcCommand(data []string, rawLine []byte) {

}

func (c ClientCommand) HandleCommand(data []string, rawLine []byte) (*Result, error) {
	switch c {
	case AddAtc:
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
		return ResultSuccess(), nil
	}
	return ResultSuccess(), nil
}
