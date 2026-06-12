package guac

type Status int

const (
	Undefined Status = -1

	Success Status = iota

	Unsupported

	ServerError

	ServerBusy

	UpstreamTimeout

	UpstreamError

	ResourceNotFound

	ResourceConflict

	ResourceClosed

	UpstreamNotFound

	UpstreamUnavailable

	SessionConflict

	SessionTimeout

	SessionClosed

	ClientBadRequest

	ClientUnauthorized

	ClientForbidden

	ClientTimeout

	ClientOverrun

	ClientBadType

	ClientTooMany
)

type statusData struct {
	name          string
	httpCode      int
	websocketCode int
	guacCode      int
}

func newStatusData(name string, httpCode, websocketCode, guacCode int) (ret statusData) {
	ret.name = name
	ret.httpCode = httpCode
	ret.websocketCode = websocketCode
	ret.guacCode = guacCode
	return
}

var guacamoleStatusMap = map[Status]statusData{
	Success:             newStatusData("Success", 200, 1000, 0x0000),
	Unsupported:         newStatusData("Unsupported", 501, 1011, 0x0100),
	ServerError:         newStatusData("SERVER_ERROR", 500, 1011, 0x0200),
	ServerBusy:          newStatusData("SERVER_BUSY", 503, 1008, 0x0201),
	UpstreamTimeout:     newStatusData("UPSTREAM_TIMEOUT", 504, 1011, 0x0202),
	UpstreamError:       newStatusData("UPSTREAM_ERROR", 502, 1011, 0x0203),
	ResourceNotFound:    newStatusData("RESOURCE_NOT_FOUND", 404, 1002, 0x0204),
	ResourceConflict:    newStatusData("RESOURCE_CONFLICT", 409, 1008, 0x0205),
	ResourceClosed:      newStatusData("RESOURCE_CLOSED", 404, 1002, 0x0206),
	UpstreamNotFound:    newStatusData("UPSTREAM_NOT_FOUND", 502, 1011, 0x0207),
	UpstreamUnavailable: newStatusData("UPSTREAM_UNAVAILABLE", 502, 1011, 0x0208),
	SessionConflict:     newStatusData("SESSION_CONFLICT", 409, 1008, 0x0209),
	SessionTimeout:      newStatusData("SESSION_TIMEOUT", 408, 1002, 0x020A),
	SessionClosed:       newStatusData("SESSION_CLOSED", 404, 1002, 0x020B),
	ClientBadRequest:    newStatusData("CLIENT_BAD_REQUEST", 400, 1002, 0x0300),
	ClientUnauthorized:  newStatusData("CLIENT_UNAUTHORIZED", 403, 1008, 0x0301),
	ClientForbidden:     newStatusData("CLIENT_FORBIDDEN", 403, 1008, 0x0303),
	ClientTimeout:       newStatusData("CLIENT_TIMEOUT", 408, 1002, 0x0308),
	ClientOverrun:       newStatusData("CLIENT_OVERRUN", 413, 1009, 0x030D),
	ClientBadType:       newStatusData("CLIENT_BAD_TYPE", 415, 1003, 0x030F),
	ClientTooMany:       newStatusData("CLIENT_TOO_MANY", 429, 1008, 0x031D),
}

func (s Status) String() string {
	if v, ok := guacamoleStatusMap[s]; ok {
		return v.name
	}
	return ""
}

func (s Status) GetHTTPStatusCode() int {
	if v, ok := guacamoleStatusMap[s]; ok {
		return v.httpCode
	}
	return -1
}

func (s Status) GetWebSocketCode() int {
	if v, ok := guacamoleStatusMap[s]; ok {
		return v.websocketCode
	}
	return -1
}

func (s Status) GetGuacamoleStatusCode() int {
	if v, ok := guacamoleStatusMap[s]; ok {
		return v.guacCode
	}
	return -1
}

func FromGuacamoleStatusCode(code int) (ret Status) {
	for k, v := range guacamoleStatusMap {
		if v.guacCode == code {
			ret = k
			return
		}
	}
	ret = Undefined
	return

}
