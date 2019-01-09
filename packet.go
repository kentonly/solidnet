package solidnet

import (
	"encoding/binary"
)

/**********************数据包interface**********************/
type IPacket interface {
	GetTotalLen() int32
	GetBodyLen() int32
	GetHeadLen() int32
	//GetCmd() int32

	GetData() []byte
	Copy(Data []byte)
	Refer(Data []byte)

	ReadByte() byte
	ReadString() string
	ReadInt16() int16
	ReadInt32() int32
	ReadInt64() int64
	ReadInt16B() int16
	ReadInt32B() int32
	ReadInt64B() int64

	WriteByte(value byte)
	WriteBytes(value []byte)
	WriteString(value string)
	WriteInt16(value int16)
	WriteInt32(value int32)
	WriteInt64(value int64)
	WriteInt16B(value int16)
	WriteInt32B(value int32)
	WriteInt64B(value int64)
}

type IPacketFactory interface {
	NewPacket() IPacket
}

/**********************基本数据包的属性和方法实现**********************/
type BasePacket struct {
	Data         []byte //报数据缓冲区
	Index        int32  //写数据索引游标
	HeadLen      int32  //包头长度
	BodyLenIndex int32  //包体长度字段起始位置，默认2字节
}

func (p *BasePacket) GetTotalLen() int32 {
	return int32(len(p.Data))
}

func (p *BasePacket) GetData() []byte {
	return p.Data
}

func (p *BasePacket) GetBodyLen() int32 {
	return int32(binary.LittleEndian.Uint16(p.Data[p.BodyLenIndex:]))
}

func (p *BasePacket) GetHeadLen() int32 {
	return p.HeadLen
}

func (p *BasePacket) Copy(Data []byte) {
	p.Data = append(p.Data[0:], Data...)
	p.Index += p.GetHeadLen()
}

func (p *BasePacket) Refer(Data []byte) {
	p.Data = Data
	p.Index = p.GetHeadLen()
}

/**********************基本数据读取**********************/
func (p *BasePacket) ReadByte() byte {
	var value byte = byte(0)
	if p.Index < int32(len(p.Data)) {
		value = p.Data[p.Index]
		p.Index++
	}
	return value
}

func (p *BasePacket) ReadString() string {
	value := string("")
	if p.Index+4 <= int32(len(p.Data)) {
		strLen := int32(binary.LittleEndian.Uint32(p.Data[p.Index:]))
		p.Index += 4
		if p.Index+strLen <= int32(len(p.Data)) {
			//包中的字符串是'\0'结尾，在C/C++中不会有任何问题，但是golang的string内部
			//都是字节序，'\0'并无特殊对待，作为正常的字符处理，这里需要剔除末尾的'\0'
			if 0 == p.Data[p.Index+strLen-1] {
				value = string(p.Data[p.Index : p.Index+strLen-1])
			} else {
				value = string(p.Data[p.Index : p.Index+strLen])
			}
			p.Index += strLen
		}
	}
	return value
}

//小端读取
func (p *BasePacket) ReadInt16() int16 {
	var value int16 = -1
	if p.Index+2 <= int32(len(p.Data)) {
		value = int16(binary.LittleEndian.Uint16(p.Data[p.Index:]))
		p.Index += 2
	}
	return value
}

func (p *BasePacket) ReadInt32() int32 {
	var value int32 = -1
	if p.Index+4 <= int32(len(p.Data)) {
		value = int32(binary.LittleEndian.Uint32(p.Data[p.Index:]))
		p.Index += 4
	}
	return value
}

func (p *BasePacket) ReadInt64() int64 {
	var value int64 = -1
	if p.Index+8 <= int32(len(p.Data)) {
		value = int64(binary.LittleEndian.Uint64(p.Data[p.Index:]))
		p.Index += 8
	}
	return value
}

//大端读取
func (p *BasePacket) ReadInt16B() int16 {
	var value int16 = -1
	if p.Index+2 <= int32(len(p.Data)) {
		value = int16(binary.BigEndian.Uint16(p.Data[p.Index:]))
		p.Index += 2
	}
	return value
}

func (p *BasePacket) ReadInt32B() int32 {
	var value int32 = -1
	if p.Index+4 <= int32(len(p.Data)) {
		value = int32(binary.BigEndian.Uint32(p.Data[p.Index:]))
		p.Index += 4
	}
	return value
}

func (p *BasePacket) ReadInt64B() int64 {
	var value int64 = -1
	if p.Index+8 <= int32(len(p.Data)) {
		value = int64(binary.BigEndian.Uint64(p.Data[p.Index:]))
		p.Index += 8
	}
	return value
}

/**********************基本数据写入**********************/
func (p *BasePacket) WriteByte(value byte) {
	p.Data = append(p.Data, value)
}

func (p *BasePacket) WriteBytes(value []byte) {
	p.Data = append(p.Data, value...)
}

func (p *BasePacket) WriteString(value string) {
	s := []byte(value)
	s = append(s, 0)
	p.WriteInt32(int32(len(s)))
	p.WriteBytes(s)
}

//小端写入
func (p *BasePacket) WriteInt16(value int16) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf[0:], uint16(value))
	p.Data = append(p.Data, buf...)
}

func (p *BasePacket) WriteInt32(value int32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf[0:], uint32(value))
	p.Data = append(p.Data, buf...)
}

func (p *BasePacket) WriteInt64(value int64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf[0:], uint64(value))
	p.Data = append(p.Data, buf...)
}

//大端写入
func (p *BasePacket) WriteInt16B(value int16) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf[0:], uint16(value))
	p.Data = append(p.Data, buf...)
}

func (p *BasePacket) WriteInt32B(value int32) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:], uint32(value))
	p.Data = append(p.Data, buf...)
}

func (p *BasePacket) WriteInt64B(value int64) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf[0:], uint64(value))
	p.Data = append(p.Data, buf...)
}
