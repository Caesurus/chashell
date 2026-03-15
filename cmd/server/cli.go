package main

import (
	"chashell/lib/transport"
	"fmt"
	"github.com/c-bata/go-prompt"
	"os"
	"strings"
)

func queueCommandForSession(sessionID, command string) {
	initPacket, dataPackets := transport.Encode([]byte(command+"\n"), true, encryptionKey, targetDomain, nil)
	_, valid := packetQueue[sessionID]
	if !valid {
		packetQueue[sessionID] = make([]string, 0)
	}
	packetQueue[sessionID] = append(packetQueue[sessionID], initPacket)
	for _, packet := range dataPackets {
		packetQueue[sessionID] = append(packetQueue[sessionID], packet)
	}
}

var commands = []prompt.Suggest{
	{Text: "sessions", Description: "Interact with the specified machine."},
	{Text: "exit", Description: "Stop the Chashell Server"},
}

func Completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}
	args := strings.Split(d.TextBeforeCursor(), " ")

	return argumentsCompleter(args)
}

func argumentsCompleter(args []string) []prompt.Suggest {
	// When interacting with a session, only basic controls are available.
	if currentSession != "" {
		return prompt.FilterHasPrefix(
			[]prompt.Suggest{
				{Text: "background", Description: "Return to main prompt"},
			},
			args[len(args)-1],
			true,
		)
	}

	if len(args) <= 1 {
		return prompt.FilterHasPrefix(commands, args[0], true)
	}

	first := args[0]
	switch first {
	case "sessions":
		second := args[1]
		if len(args) == 2 {
			sessions := []prompt.Suggest{}
			for clientGUID, clientInfo := range sessionsMap {
				sessions = append(sessions, prompt.Suggest{Text: clientGUID, Description: clientInfo.hostname})
			}

			return prompt.FilterHasPrefix(sessions, second, true)
		}

	}
	return []prompt.Suggest{}
}

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	// If we are currently interacting with a session, all input except
	// the local "background" control is forwarded to the remote shell.
	if currentSession != "" {
		if in == "background" {
			fmt.Println("Returning to main prompt.")
			currentSession = ""
			return
		}
		queueCommandForSession(currentSession, in)
		return
	}

	args := strings.Split(in, " ")
	if len(args) > 0 {
		switch args[0] {
		case "exit":
			fmt.Println("Exiting.")
			os.Exit(0)
		case "sessions":
			if len(args) == 2 {
				sessionID := args[1]
				fmt.Printf("Interacting with session %s.\n", sessionID)

				// Print any buffered data we have for this session and clear it.
				buffer, dataAvailable := consoleBuffer[sessionID]
				if dataAvailable && buffer.Len() > 0 {
					fmt.Println(buffer.String())
				}
				delete(consoleBuffer, sessionID)

				currentSession = sessionID
			} else {
				fmt.Println("sessions [id]")
			}
		}
	}
}
