package packet

import "math"

type BroadcastFilter func(toClient, fromClient *Client) bool

func BroadcastToAll(_, _ *Client) bool {
	return true
}

func BroadcastToAtc(toClient, _ *Client) bool {
	return toClient.IsAtc
}

func BroadcastToClientInRange(toClient, fromClient *Client) bool {
	if fromClient == nil {
		return true
	}
	distance := FindNearestDistance(toClient.Position, fromClient.Position)
	var threshold float64 = 0
	switch {
	case toClient.IsAtc && fromClient.IsAtc:
		threshold = math.Max(toClient.VisualRange, fromClient.VisualRange)
	case toClient.IsAtc:
		threshold = toClient.VisualRange
	case fromClient.IsAtc:
		threshold = fromClient.VisualRange
	default:
		threshold = toClient.VisualRange + fromClient.VisualRange
	}
	return distance <= threshold
}

func CombineBroadcastFilter(filters ...BroadcastFilter) BroadcastFilter {
	return func(toClient, fromClient *Client) bool {
		for _, f := range filters {
			if f == nil {
				continue
			}
			if !f(toClient, fromClient) {
				return false
			}
		}
		return true
	}
}
