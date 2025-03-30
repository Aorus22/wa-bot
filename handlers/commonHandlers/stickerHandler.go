package commonHandlers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"wa-bot/utils"
)

func StickerHandler(
	client *whatsmeow.Client,
	senderJID waTypes.JID,
	vMessage *waProto.Message,
	messageText string,
) {
	client.SendMessage(context.Background(), senderJID, &waProto.Message{
		Conversation: proto.String("⏳ Loading..."),
	})

	crop := strings.Contains(strings.ToLower(messageText), "crop")

	var (
		mediaPath string
		isVideo   bool
		err       error
	)

	switch {
	case vMessage.GetImageMessage() != nil:
		mediaPath, isVideo, err = getWaMedia(client, vMessage.GetImageMessage(), false)
	case vMessage.GetVideoMessage() != nil:
		mediaPath, isVideo, err = getWaMedia(client, vMessage.GetVideoMessage(), true)
	default:
		mediaPath, isVideo, err = getMediaFromUrl(messageText)
	}

	if err != nil {
		fmt.Println("Failed to process media:", err)
		return
	}

	sendMediaAsSticker(client, senderJID, mediaPath, crop, isVideo)
}

func getWaMedia(
	client *whatsmeow.Client,
	media whatsmeow.DownloadableMessage,
	isVideo bool,
) (string, bool, error) {
	data, err := client.Download(media)
	if err != nil {
		return "", false, fmt.Errorf("download failed: %w", err)
	}

	ext := ".jpg"
	if isVideo {
		ext = ".mp4"
	}
	mediaPath := fmt.Sprintf("media/%d%s", time.Now().UnixMilli(), ext)

	err = os.WriteFile(mediaPath, data, 0644)
	if err != nil {
		return "", false, fmt.Errorf("failed to save media: %w", err)
	}

	return mediaPath, isVideo, nil
}

func getMediaFromUrl(messageText string) (string, bool, error) {
	url, err := utils.GetLinkFromString(messageText)
	if err != nil {
		return "", false, fmt.Errorf("invalid URL: %w", err)
	}

	mediaPath, err := utils.DownloadMediaFromURL(url)
	if err != nil {
		return "", false, fmt.Errorf("failed to download from URL: %w", err)
	}

	mimeType, err := utils.GetMimeType(mediaPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to get MIME type: %w", err)
	}

	isVideo := strings.HasPrefix(mimeType, "video/")
	return mediaPath, isVideo, nil
}

func sendMediaAsSticker(client *whatsmeow.Client, senderJID waTypes.JID, mediaPath string, crop bool, isAnimated bool) {
	webpPath, err := utils.ConvertToWebp(mediaPath, crop)
	if err != nil {
		fmt.Println("Convert failed:", err)
		return
	}
	defer os.Remove(webpPath)

	webpData, err := os.ReadFile(webpPath)
	if err != nil {
		fmt.Println("Failed to read WebP file:", err)
		return
	}

	uploaded, err := client.Upload(context.Background(), webpData, whatsmeow.MediaImage)
	if err != nil {
		fmt.Println("Failed to upload sticker:", err)
		return
	}

	_, err = client.SendMessage(context.Background(), senderJID, &waProto.Message{
		StickerMessage: &waProto.StickerMessage{
			Mimetype:      proto.String("image/webp"),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			IsAnimated:    proto.Bool(isAnimated),
		},
	})

	if err != nil {
		client.SendMessage(context.Background(), senderJID, &waProto.Message{
			Conversation: proto.String("Gagal dalam membuat sticker"),
		})
	}
}
