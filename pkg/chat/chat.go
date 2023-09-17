package chat

import (
	"errors"
	"flag"
	"strings"

	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/isaacpd/costanza/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ChatAddress = flag.String("chat", "localhost:8000", "The address of the chat service")

	client proto.ChatServiceClient
)

func Init() error {
	conn, err := grpc.Dial(*ChatAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	client = proto.NewChatServiceClient(conn)
	return nil
}

func HandleChat(c cmd.Context) (string, error) {
	var sent proto.ChatMessage
	sent.User = c.Author.Username
	sent.Content = strings.ReplaceAll(c.Message.Content, c.Session.State.User.Mention(), "Costanza")
	if client == nil {
		return "", errors.New("client not initialized")
	}

	stream, err := client.Chat(c.Context)
	if err != nil {
		return "", err
	}
	stream.Send(&sent)

	received, err := stream.Recv()
	stream.CloseSend()
	if err != nil {
		return "", err
	}
	return received.Content, nil
}
