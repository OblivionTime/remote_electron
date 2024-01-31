package turn

import (
	"fmt"
	"net"
	"remote_server/config"
	"remote_server/global"
	"remote_server/model"

	"github.com/pion/turn/v3"
)

func NewServer() error {
	udpListener, err := net.ListenPacket("udp4", fmt.Sprintf("0.0.0.0:%d", config.Config.Turn.Port))

	if err != nil {
		fmt.Println(" net.ResolveUDPAddr error", err)
		return err
	}
	global.Turnserver = model.NewTurnServer()

	s, err := turn.NewServer(turn.ServerConfig{
		Realm:       "oblivionTime",
		AuthHandler: global.Turnserver.Authenticate,
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(config.Config.Turn.PublicIP), // Claim that we are listening on IP passed by user (This should be your Public IP)
					Address:      "0.0.0.0",                                // But actually be listening on every interface
				},
			},
		},
	})
	if err != nil {
		fmt.Println("Failed to create TURN server:", err)
		return err
	}
	global.Turnserver.Svr = s
	return nil
}
