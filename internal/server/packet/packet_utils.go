package packet

import (
	"bytes"
	"strings"
)

func parserCommandLine(line []byte) (ClientCommand, []string) {
	for _, prefix := range PossibleClientCommands {
		if bytes.HasPrefix(line, prefix) {
			decodeLine := string(line[len(prefix):])
			return prefix, strings.Split(decodeLine, ":")
		}
	}
	return nil, nil
}

func makePacket(command ClientCommand, parts ...string) []byte {
	totalLen := len(command)
	if len(parts) > 0 {
		for _, part := range parts {
			totalLen += len(part)
		}
		totalLen += len(parts) - 1
	}

	result := make([]byte, totalLen)
	pos := 0

	pos += copy(result[pos:], command)

	for i, part := range parts {
		if i > 0 {
			result[pos] = ':'
			pos++
		}
		pos += copy(result[pos:], part)
	}

	return result
}
