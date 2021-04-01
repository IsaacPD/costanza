package util

import "strings"

var (
	ArgIdentifier = " "
)

func AfterCommand(message string) string {
	return message[strings.Index(message, ArgIdentifier)+1:]
}
