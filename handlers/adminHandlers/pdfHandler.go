package adminHandlers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"wa-bot/utils"
)

func SendPDFHandler(client *whatsmeow.Client, isFromGroup bool, senderJID waTypes.JID, vMessage *waProto.Message, messageText string) {

	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String("⏳ Loading..."),
	})

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

	messageArray := strings.Split(messageText, " ")
	if len(messageArray) < 2 && len(messageArray) > 3 {
		return
	}

	mapel := messageArray[1]
	listMapel, err := utils.FetchMapel()
	if err != nil {
		client.SendMessage(context.Background(), senderJID, &waProto.Message{
			Conversation: proto.String("Gagal mengambil daftar mapel."),
		})
		return
	}

	if index, err := strconv.Atoi(mapel); err == nil {
		if index > 0 && index <= len(listMapel) {
			mapel = listMapel[index-1]
		} else {
			client.SendMessage(context.Background(), senderJID, &waProto.Message{
				Conversation: proto.String("Nomor mapel tidak valid."),
			})
			return
		}
	} else if !utils.Contains(listMapel, mapel) {
		client.SendMessage(context.Background(), senderJID, &waProto.Message{
			Conversation: proto.String("Mapel tidak valid."),
		})
		return
	}

	if len(messageArray) == 3 {
		answer := messageArray[2]
		sendPDFMessage(client, senderJID, mapel, answer)
	} else if len(messageArray) == 2 {
		sendPDFMessage(client, senderJID, mapel, "")
	} else {
		client.SendMessage(context.Background(), senderJID, &waProto.Message{
			Conversation: proto.String("Invalid Command"),
		})
	}
}

func sendPDFMessage(client *whatsmeow.Client, senderJID waTypes.JID, mapel string, answer string) {
	var pdfPath string
	var err error

	if answer == "" {
		pdfPath, err = utils.FetchPDF(mapel)
	} else {
		jsonAnswer, err := convertToJSON(answer)

		if err != nil {
			client.SendMessage(context.Background(), senderJID, &waProto.Message{
				Conversation: proto.String("Format Jawaban Salah"),
			})
		}
		pdfPath, _ = utils.FetchPDF(mapel, jsonAnswer)
	}

	if err != nil {
		fmt.Println("Failed to fetch PDF:", err)
		client.SendMessage(context.Background(), senderJID, &waProto.Message{
			Conversation: proto.String("Gagal mengambil PDF"),
		})
		return
	}

	fileData, err := os.ReadFile(pdfPath)
	if err != nil {
		fmt.Println("Failed to read PDF file:", err)
		return
	}

	uploaded, err := client.Upload(context.Background(), fileData, whatsmeow.MediaDocument)
	if err != nil {
		fmt.Println("Failed to upload PDF:", err)
		return
	}

	_, err = client.SendMessage(context.Background(), senderJID, &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			Title:        proto.String(mapel),
			Mimetype:     proto.String("application/pdf"),
			URL:          proto.String(uploaded.URL),
			DirectPath:   proto.String(uploaded.DirectPath),
			MediaKey:     uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:   uploaded.FileSHA256,
			FileLength:   proto.Uint64(uploaded.FileLength),
		},
	})
	if err != nil {
		fmt.Println("Failed to send PDF:", err)
	}

	err = os.Remove(pdfPath)
	if err != nil {
		fmt.Println("Failed to delete PDF file:", err)
	}
}

func convertToJSON(input string) (map[string]string, error) {
	lines := strings.Split(input, "\n")

	dataKunci := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "-" || line == "" {
			continue
		}

		parts := strings.SplitN(line, ".", 2)
		if len(parts) == 2 {
			nomor := strings.TrimSpace(parts[0])
			jawaban := strings.TrimSpace(parts[1])
			dataKunci[nomor] = strings.ToUpper(jawaban)
		}
	}

	return dataKunci, nil
}