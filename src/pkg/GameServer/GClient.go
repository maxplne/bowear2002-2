package GameServer

import (
	C "Core"
	"log"
	D "Data"
	//R "reflect"
)

type GClient struct {
	C.Client
	Key    byte
	packet *C.Packet

	ID     uint32
	Player *D.Player
	Server *GServer
	Map    *Map
}

func (client *GClient) StartRecive() {
	defer client.OnDisconnect()
	for {
		bl, err := client.Socket.Read(client.packet.Buffer[client.packet.Index:])
		if err != nil {
			return
		}

		client.packet.Index += bl

		for client.packet.Index > 2 {
			p := client.packet
			size := p.Index
			p.Index = 0
			if p.ReadByte() != 0xAA { 
				client.Log().Printf("Wrong packet header")
				client.Log().Printf("% #X", p.Buffer[:size])
				return
			} 
			l := int(p.ReadUInt16())
			p.Index = size
			if len(client.packet.Buffer) < l {
				client.packet.Resize(l)
			} 

			if size >= l+3 {
				temp := client.packet.Buffer[:l+3]
				op := client.packet.Buffer[3]
				if op > 13 || (op > 1 && op < 5) || (op > 6 && op < 13) {
					var sumCheck bool
					temp, sumCheck = C.DecryptPacket(temp)
					if !sumCheck {
						client.Log().Println("Packet sum check failed!")
						return
					}
				} else {
					temp = temp[3:]
				}
				client.ParsePacket(C.NewPacketRef(temp))
				client.packet.Index = 0
				if size > l+3 {
					client.packet.Index = size - (l + 3)
					copy(client.packet.Buffer, client.packet.Buffer[l+3:size])
				} else {
					//keeping the user under 4048k use to save memory
					if cap(client.packet.Buffer) > 4048 {
						client.Buffer = make([]byte, 1024)
						client.packet = C.NewPacketRef(client.Buffer) 
					}
				}
			} else {
				break
			}
		}
	}
}


func (client *GClient) OnConnect() {

	userID, q := D.LoginQueue.Check(client.IP)
	if !q {
		client.OnDisconnect()
		return
	}

	id, r := client.Server.IDG.Next()

	if !r {
		client.OnDisconnect()
		return
	}

	client.Log().Println("ID " + userID)
	client.Player = D.GetPlayerByUserID(userID)
	client.Log().Println("name " + client.Player.Name)

	client.packet = C.NewPacketRef(client.Buffer)
	client.packet.Index = 0
	client.ID = id

	Server.Run.Funcs <- func() {
		client.Server.Maps[0].OnPlayerJoin(client)
		client.SendWelcome()
	}
	client.StartRecive()
}


func (client *GClient) OnDisconnect() {
	if x := recover(); x != nil {
		client.Log().Printf("panic : %v",x)
	} 
	if client.Map != nil {
		client.Map.OnLeave(client)
	}
	if client.Player != nil {
		client.Server.IDG.Return(client.ID) 
		client.Server.DBRun.Funcs <- func() { D.SavePlayer(client.Player)  }
	}
	client.Socket.Close()
	client.MainServer.GetServer().Log.Println("Client Disconnected!")
}

func (client *GClient) Send(p *C.Packet) {
	if !p.Encrypted {
		op := p.Buffer[3]
		if op > 13 || (op > 0 && op < 3) || (op > 3 && op < 11) {
			p.WSkip(2)
			C.EncryptPacket(p.Buffer[:p.Index], client.Key)
			p.Encrypted = true
			client.Key++
		}
		p.WriteLen()
	}
	client.Socket.Write(p.Buffer[:p.Index])
}

func (client *GClient) SendRaw(p *C.Packet) {
	p.WriteLen()
	client.Socket.Write(p.Buffer[:p.Index])
} 
 

func (client *GClient) SendWelcome() {

	//player stats
	packet := C.NewPacket2(77)
	packet.WriteHeader(0x15)
	packet.WriteUInt32(client.ID)
	packet.Write([]byte{0x00, 0x00, 0x00, 0x0C, 0x00, 0x00, 0x00, 0x0C, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x93, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x05, 0x05, 0x05, 0x05, 0x30, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x01, 0x00, 0x00, 0x01, 0x19, 0x00})
	//packet.Write([]byte{0x00, 0x00, 0x00, 0x0C, 0x00, 0x00, 0x00, 0x0C, 0x07, 0x00, 0x00, 0x00, 0x01, 0x00, 0x04, 0x95, 0xD4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x0F, 0x0A, 0x0A, 0x05, 0x30, 0x00, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x64, 0x00, 0x01, 0x00, 0x00, 0x01, 0x19, 0x00})
	client.Send(packet)
  
	//send map info
	packet = C.NewPacket2(198)
	packet.WriteHeader(0x17)
	packet.Write([]byte{0x00, 0x01, 0x87, 0x0A, 0x01, 0x00, 0x00, 0x00, 0x0C, 0x00, 0x00, 0x7B, 0x48, 0x98, 0xE4, 0x7B, 0x49, 0xF8, 0x74, 0x00, 0x0D, 0x00}) //, 0x00, 0x00, 0x0D, 0x42, 0x01, 0xC0, 0x0C, 0x40, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x41, 0x0C, 0x00, 0x0B, 0x40, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x40, 0x04, 0x00, 0x08, 0x80, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x3F, 0x0A, 0x60, 0x07, 0xA0, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x3E, 0x03, 0xA0, 0x02, 0x80, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x3D, 0x0B, 0xE0, 0x02, 0x40, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x3C, 0x07, 0x80, 0x01, 0x60, 0x1D, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x00, 0x00, 0x0D, 0x3B, 0x06, 0x20, 0x04, 0x40, 0x1D, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x00, 0x00, 0x0D, 0x3A, 0x07, 0x40, 0x08, 0x80, 0x1D, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x00, 0x00, 0x0D, 0x39, 0x09, 0x40, 0x0C, 0x00, 0x1D, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x00, 0x00, 0x0D, 0x38, 0x07, 0xC0, 0x0E, 0xA0, 0x1D, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x0D})
	client.Send(packet)
	
	//packet = C.NewPacket2(243)
	//packet.WriteHeader(0x17)
	//packet.Write([]byte{0x00, 0x01, 0x87, 0x0B, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x00, 0x01, 0x69, 0x28, 0x5B, 0xE8, 0x69, 0x29, 0xBB, 0x78, 0x00, 0x0C, 0x0E, 0x00, 0x00, 0x0C, 0x56, 0x0D, 0x40, 0x0D, 0xA0, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x55, 0x08, 0x20, 0x0C, 0x60, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x54, 0x03, 0xA0, 0x0B, 0x80, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x53, 0x0A, 0x60, 0x09, 0xC0, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x52, 0x05, 0xE0, 0x09, 0xC0, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x51, 0x04, 0x20, 0x04, 0xC0, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x50, 0x0C, 0xC0, 0x04, 0x40, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x4F, 0x08, 0xC0, 0x03, 0xC0, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x4E, 0x02, 0xA0, 0x02, 0x80, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x4D, 0x0D, 0x60, 0x02, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x4C, 0x08, 0xC0, 0x05, 0x60, 0x17, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x4B, 0x05, 0x60, 0x07, 0xC0, 0x17, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x4A, 0x0B, 0x40, 0x08, 0x00, 0x17, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x08, 0x80, 0x0A, 0xA0, 0x17, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00})
	//client.Send(packet)
	

	//send player name 
	packet = C.NewPacket2(28 + len(client.Map.Players)*13)
	packet.WriteHeader(0x47)
	packet.WriteInt16(int16(len(client.Map.Players)))
	for _, s := range client.Map.Players {
		packet.WriteString(s.Player.Name)
		packet.WSkip(2)
	}
	client.Send(packet)

	//send spawn palyer
	client.Map.OnPlayerAppear(client)

	packet = C.NewPacket2(18)
	packet.WriteHeader(0x0E)
	packet.Write([]byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	client.Send(packet)

	packet = C.NewPacket2(13)
	packet.WriteHeader(0x0E)
	packet.Write([]byte{0x05, 0x00})
	client.Send(packet) 

	packet = C.NewPacket2(21)
	packet.WriteHeader(0x0E)
	packet.Write([]byte{0x01, 0x0E, 0x14, 0xFC, 0x50, 0x05, 0x20, 0xB7, 0x2B, 0x00})
	client.Send(packet)

	//send ready to get ingame
	packet = C.NewPacket2(17)
	packet.WriteHeader(0x0E)
	packet.Write([]byte{0x04, 0x01, 0x05, 0x20, 0xB7, 0x46})
	client.Send(packet)

	packet = C.NewPacket2(13)
	packet.WriteHeader(0x3E)
	packet.Write([]byte{0x00, 0x00})
	//client.Map.Send(packet)
}

func (client *GClient) Log() *log.Logger {
	return Server.Log
} 

func (client *GClient) ParsePacket(p *C.Packet) {
	header := p.ReadByte()

	fnc, exist := Handler[int(header)]
	if !exist {
		client.Log().Printf("isnt registred : %s",p)
		return
	}
	//client.Log().Printf("Header(%d) len(%d) : % #X\n %s", header, len(p.Buffer), p.Buffer, p.Buffer)
	//client.Log().Printf("Handle %s\n", R.TypeOf(fnc))
	
	fnc(client, p)
}
