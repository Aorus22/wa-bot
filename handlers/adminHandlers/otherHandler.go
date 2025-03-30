package adminHandlers

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"wa-bot/utils"
)

func ListgroupsHandler(client *whatsmeow.Client, senderJID waTypes.JID){
	allowedSender := os.Getenv("ALLOWED_SENDER")
	if senderJID.String() != allowedSender {
		fmt.Println("Sender not allowed for !listgroups command.")
		return
	}

	groups, err := client.GetJoinedGroups()
	if err != nil {
		fmt.Println("Error fetching joined groups:", err)
		return
	}

	responseText := "📌 *Daftar Grup:*\n\n"
	for _, group := range groups {
		responseText += fmt.Sprintf("📂 *%s*\n📎 ID: %s\n", group.Name, group.JID.String())

		_, err := client.GetGroupInfo(group.JID)
		if err != nil {
			fmt.Println("Failed to get group info for", group.JID.String(), ":", err)
			continue
		}
	}

	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String(responseText),
	})
}

func ListMapelHandler(client *whatsmeow.Client, isFromGroup bool, senderJID waTypes.JID) {
	groupJIDs := strings.Split(os.Getenv("GROUP_JIDS"), ",")

	userRegistered := false

	for _, groupJIDStr := range groupJIDs {
		targetGroupJID, err := waTypes.ParseJID(groupJIDStr)
		if err != nil {
			fmt.Println("Invalid group JID:", err)
			continue
		}

		groupInfo, err := client.GetGroupInfo(targetGroupJID)
		if err != nil {
			fmt.Println("Failed to get group info for", groupJIDStr, ":", err)
			continue
		}

		for _, participant := range groupInfo.Participants {
			if participant.JID.String() == senderJID.String() {
				userRegistered = true
				break
			}
		}

		if userRegistered {
			break
		}
	}

	if !isFromGroup && !userRegistered {
		return
	}

	listMapel, err := utils.FetchMapel()
	if err != nil {
		fmt.Println("Failed to fetch mapel:", err)
		return
	}

	var listMapelString string
	for i, mapel := range listMapel {
		listMapelString += fmt.Sprintf("%d. %s\n", i+1, mapel)
	}

	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String(listMapelString),
	})
}
