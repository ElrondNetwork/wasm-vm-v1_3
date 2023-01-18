package common

import (
	"os"

	"github.com/multiversx/mx-chain-vm-v1_3-go/vmhost"
	"github.com/multiversx/mx-chain-vm-v1_3-go/ipc/marshaling"
)

// ArwenArguments represents the initialization arguments required by Arwen, passed through the initialization pipe
type ArwenArguments struct {
	vmhost.VMHostParameters
	LogsMarshalizer     marshaling.MarshalizerKind
	MessagesMarshalizer marshaling.MarshalizerKind
}

// SendArwenArguments sends initialization arguments through a pipe
func SendArwenArguments(pipe *os.File, pipeArguments ArwenArguments) error {
	sender := NewSender(pipe, createArgumentsMarshalizer())
	message := NewMessageInitialize(pipeArguments)
	_, err := sender.Send(message)
	return err
}

// GetArwenArguments reads initialization arguments from the pipe
func GetArwenArguments(pipe *os.File) (*ArwenArguments, error) {
	receiver := NewReceiver(pipe, createArgumentsMarshalizer())
	message, _, err := receiver.Receive(0)
	if err != nil {
		return nil, err
	}

	typedMessage := message.(*MessageInitialize)
	return &typedMessage.Arguments, nil
}

// For the arguments, the marshalizer is fixed to JSON
func createArgumentsMarshalizer() marshaling.Marshalizer {
	return marshaling.CreateMarshalizer(marshaling.JSON)
}
