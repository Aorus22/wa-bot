package adminHandlers

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"wa-bot/utils"
)

func TokenHandler(client *whatsmeow.Client, senderJID waTypes.JID) {
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

	if !userRegistered {
		return
	}

	utils.UserState.AddUser(senderJID.String())

	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String("Silakan masukkan nama lengkap Anda."),
	})
}

func GetNameHandler(client *whatsmeow.Client, senderJID waTypes.JID, messageText string, startTime time.Time) {

	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String("⏳ Loading..."),
	})

	timeoutStr := os.Getenv("TIMEOUT_NAMA")

	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = 2
	}

	if time.Since(startTime) > time.Duration(timeout)*time.Minute {
		utils.UserState.ClearUser(senderJID.String())

		client.SendMessage(context.Background(), senderJID, &waProto.Message{
			Conversation: proto.String("⏳ Waktu habis! Silakan ketik *!token* lagi."),
		})

		return
	}

	var validNameRegex = regexp.MustCompile(`^[a-zA-Z' ]+$`)
	fmt.Println(messageText)
	if !validNameRegex.MatchString(messageText) {
		client.SendMessage(context.Background(), senderJID, &waProto.Message{
			Conversation: proto.String("⚠️ Nama Invalid"),
		})

		utils.UserState.ClearUser(senderJID.String())
		return
	}

	nis := strings.Split(senderJID.String(), "@")[0]
	nama := messageText

	utils.UserState.ClearUser(senderJID.String())

	status, token, err := utils.FetchTokenData(nama, nis)
	if err != nil {
		fmt.Println("Failed to fetch token:", err)
		return
	}

	var responseText string
	if status == "new" {
		responseText = "✅ Token baru Anda adalah:"
	} else if status == "update" {
		responseText = "Token lama telah tidak berlaku. Ini token baru anda:"
	} else {
		responseText = "Gagal mendapatkan token."
	}

	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String(responseText),
	})

	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String(token),
	})
}
