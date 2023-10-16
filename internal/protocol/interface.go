package protocol

type IProtocol interface {
	StartRoutine(string) (error)
	DispatchMessage() (error)
	ProcessData() (error)
	CreateDataMessage() (error)
}
