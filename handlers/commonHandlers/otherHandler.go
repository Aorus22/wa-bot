package commonHandlers

import (
	"context"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func CheckHandler(client *whatsmeow.Client, senderJID waTypes.JID){
	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String("Hello, World!"),
	})
}