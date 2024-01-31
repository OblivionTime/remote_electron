package turn

import (
	"context"
	"fmt"
	"log"
	"net"
	"remote_server/config"
	"remote_server/global"
	"remote_server/model"
	"syscall"

	"github.com/pion/turn/v3"
	"golang.org/x/sys/unix"
)

func NewServer() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", config.Config.Turn.Port))
	if err != nil {
		fmt.Println(" net.ResolveUDPAddr error", err)
		return err
	}
	threadNum := config.Config.Turn.ThreadNum
	listenerConfig := &net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error
			if err = conn.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			}); err != nil {
				return err
			}

			return operr
		},
	}

	relayAddressGenerator := &turn.RelayAddressGeneratorStatic{
		RelayAddress: net.ParseIP(config.Config.Turn.PublicIP), // Claim that we are listening on IP passed by user
		Address:      "0.0.0.0",                                // But actually be listening on every interface
	}

	packetConnConfigs := make([]turn.PacketConnConfig, threadNum)
	var i uint
	for i = 0; i < threadNum; i++ {
		conn, listErr := listenerConfig.ListenPacket(context.Background(), addr.Network(), addr.String())
		if listErr != nil {
			log.Fatalf("Failed to allocate UDP listener at %s:%s", addr.Network(), addr.String())
		}

		packetConnConfigs[i] = turn.PacketConnConfig{
			PacketConn:            conn,
			RelayAddressGenerator: relayAddressGenerator,
		}

		log.Printf("Server %d listening on %s\n", i, conn.LocalAddr().String())
	}
	global.Turnserver = model.NewTurnServer()

	s, err := turn.NewServer(turn.ServerConfig{
		Realm: "oblivionTime",
		// Set AuthHandler callback
		// This is called every time a user tries to authenticate with the TURN server
		// Return the key for that user, or false when no user is found
		AuthHandler: global.Turnserver.Authenticate,
		// PacketConnConfigs is a list of UDP Listeners and the configuration around them
		PacketConnConfigs: packetConnConfigs,
	})
	if err != nil {
		fmt.Println("Failed to create TURN server:", err)
		return err
	}
	global.Turnserver.Svr = s
	return nil
}
