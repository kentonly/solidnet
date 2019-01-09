package solidnet

type IHandler interface {
	HandleTimer(IMessage)
	HandleNet(IMessage)
	HandleState(IMessage)
}
