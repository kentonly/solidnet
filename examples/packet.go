package main

import (
	"encoding/binary"

	solidnet "github.com/idakun/solidnet"
)

const (
	CLIENT_COMMAND_LOGIN_AUTH   = 0x1000 // 客户端认证请求
	SERVER_COMMAND_AUTH_SUCCESS = 0x1001 // 服务端回复认证成功
	CLIENT_COMMAND_TIME_REQ     = 0x1002 // 客户端请求当前时间
	SERVER_COMMAND_TIME_RESP    = 0x1003 // 服务端回复当前时间
)

const (
	AUTH_KEY int32 = 23458900 // 这里用一个整数作为登录验证
)

/**********************实现具体业务包**********************/
const (
	PACKET_HEADER_LEN     = 10
	PACKET_BODY_LEN_INDEX = 8
	PACKET_CMD_INDEX      = 6
	MAX_PACKET_LEN        = 20 * 1024
)

// 包头定义
//type PacketHeader struct{
//	magic [5]byte
//	ver byte
//	cmd uint16
//	bodylen uint16
//};

type Packet struct {
	solidnet.BasePacket
	CmdIndex int32 // 命令字字段起始位置，默认2字节
}

type PacketFactory struct {
}

func NewPacketFactory() *PacketFactory {
	return &PacketFactory{}
}

func (pf *PacketFactory) NewPacket() solidnet.IPacket {
	return NewPacket()
}

func NewPacket() *Packet {
	p := new(Packet)
	p.HeadLen = PACKET_HEADER_LEN
	p.BodyLenIndex = PACKET_BODY_LEN_INDEX
	p.CmdIndex = PACKET_CMD_INDEX
	p.Data = make([]byte, 0, MAX_PACKET_LEN)
	return p
}

func (p *Packet) WriteBegin(cmd int16, version byte) {
	p.WriteBytes([]byte("CHINA"))
	p.WriteByte(version)
	p.WriteInt16(cmd)
	p.WriteInt16(0)
}

func (p *Packet) WriteEnd() {
	bodyLen := int16(p.GetTotalLen() - p.GetHeadLen())
	binary.LittleEndian.PutUint16(p.Data[p.BodyLenIndex:], uint16(bodyLen))
	p.Index = p.GetHeadLen()
}

func (p *Packet) GetCmd() int32 {
	return int32(binary.LittleEndian.Uint16(p.Data[p.CmdIndex:]))
}
