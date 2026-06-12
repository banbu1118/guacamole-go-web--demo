package guac

import (
	"net"

	"github.com/sirupsen/logrus"
)

// NewGuacamoleTunnel creates a tunnel to guacd for RDP connection
func NewGuacamoleTunnel(guacadAddr, protocol, host, port, user, password, uuid string, w, h, dpi int) (s *SimpleTunnel, err error) {
	config := NewGuacamoleConfiguration()
	config.ConnectionID = uuid
	config.Protocol = protocol
	config.OptimalScreenHeight = h
	config.OptimalScreenWidth = w
	config.OptimalResolution = dpi
	config.AudioMimetypes = []string{"audio/L16", "rate=44100", "channels=2"}
	config.Parameters = map[string]string{
		"scheme":           protocol,
		"hostname":         host,
		"port":             port,
		"ignore-cert":      "true",
		"security":         "any",
		"username":         user,
		"password":         password,
		"enable-wallpaper": "true",
		"resize-method":    "display-update",
		"disable-copy":     "false",
		"disable-paste":    "false",
		"enable-audio":     "true",
	}
	addr, err := net.ResolveTCPAddr("tcp", guacadAddr)
	if err != nil {
		logrus.Errorln("error while resolving guacd address", err)
		return nil, err
	}
	logrus.Infof("Connecting to guacd at %s", addr.String())
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logrus.Errorln("error while connecting to guacd", err)
		return nil, err
	}
	logrus.Info("Connected to guacd, starting handshake")
	stream := NewStream(conn, SocketTimeout)
	err = stream.Handshake(config)
	if err != nil {
		logrus.Errorln("handshake failed", err)
		return nil, err
	}
	logrus.Info("Handshake completed successfully")
	tunnel := NewSimpleTunnel(stream)
	return tunnel, nil
}
