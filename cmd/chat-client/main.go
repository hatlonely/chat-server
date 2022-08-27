package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/hatlonely/chat-server/api/gen/go/api"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/hatlonely/go-kit/flag"
	"github.com/hatlonely/go-kit/refx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Version string

type Options struct {
	flag.Options

	Endpoint string `flag:"-e; default: 127.0.0.1:6080"`
	Username string `flag:"-u; default: hatlonely"`
	To       string `flag:"-t; default: lulu"`

	Window struct {
		Width      int `flag:"default: 50"`
		ChatHeight int `flag:"default: 20"`
		TextHeight int `flag:"default: 5"`
	}
}

func main() {
	var options Options
	refx.Must(flag.Struct(&options, refx.WithCamelName(), refx.WithDefaultValidator()))
	refx.Must(flag.Parse(flag.WithJsonVal()))
	if options.Help {
		fmt.Println(flag.Usage())
		return
	}
	if options.Version {
		fmt.Println(Version)
		return
	}

	conn, err := grpc.Dial(options.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	refx.Must(err)
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := api.NewChatServiceClient(conn)
	stream, err := client.Chat(ctx)
	refx.Must(err)
	defer stream.CloseSend()

	// termui
	refx.Must(termui.Init())
	defer termui.Close()

	// 聊天框
	var chatAreaMessages []string
	chatArea := widgets.NewList()
	chatArea.Rows = chatAreaMessages
	chatArea.SetRect(0, 0, options.Window.Width, options.Window.ChatHeight)
	termui.Render(chatArea)

	appendMessageToChatArea := func(message string) {
		chatAreaMessages = append(chatAreaMessages, message)
		chatArea.Rows = chatAreaMessages
		chatArea.ScrollBottom()
		termui.Render(chatArea)
	}

	// 输入框
	var textAreaBuffer bytes.Buffer
	textArea := widgets.NewParagraph()
	textArea.SetRect(0, options.Window.ChatHeight+1, options.Window.Width, options.Window.ChatHeight+options.Window.TextHeight+1)
	termui.Render(textArea)

	appendCharacterToTextArea := func(ch string) {
		textAreaBuffer.WriteString(ch)
		textArea.Text = textAreaBuffer.String()
		termui.Render(textArea)
	}
	clearTextArea := func() {
		textAreaBuffer.Reset()
		textArea.Text = textAreaBuffer.String()
		termui.Render(textArea, chatArea)
	}

	// send to server
	messages := make(chan string, 1)
	go func() {
	sendLoop:
		for {
			select {
			case <-ctx.Done():
				break sendLoop
			case message := <-messages:
				if err := stream.Send(&api.MessageFromClient{
					To:      options.To,
					Message: message,
				}); err != nil {
					fmt.Printf("system: %s\n", err.Error())
					continue
				}
			}
		}
	}()

	// recv from server
	go func() {
		for {
			message, err := stream.Recv()
			if err != nil {
				fmt.Printf("system: %s\n", err.Error())
				time.Sleep(time.Second)
				continue
			}

			appendMessageToChatArea(fmt.Sprintf("%s: %s", message.From, message.Message))
		}
	}()

	for e := range termui.PollEvents() {
		if e.Type == termui.KeyboardEvent {
			switch e.ID {
			case "<C-c>":
				cancel()
				return
			case "<Space>":
				appendCharacterToTextArea(" ")
			case "<Enter>":
				appendMessageToChatArea(fmt.Sprintf("%s: %s", options.Username, textAreaBuffer.String()))
				messages <- textAreaBuffer.String()
				clearTextArea()
			default:
				appendCharacterToTextArea(e.ID)
			}
		}
	}
}
