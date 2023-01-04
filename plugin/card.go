package plugin

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"

	"github.com/drone/drone-go/drone"
)

func (args Args) writeCard(data Card) error {
	result, _ := json.Marshal(data)
	card := drone.CardInput{
		Schema: "https://drone.github.io/drone-jira/card.json",
		Data:   result,
	}
	writeCard(args.CardFilePath, &card)
	return nil
}

func writeCard(path string, card interface{}) {
	data, _ := json.Marshal(card)
	switch {
	case path == "/dev/stdout":
		writeCardTo(os.Stdout, data)
	case path == "/dev/stderr":
		writeCardTo(os.Stderr, data)
	case path != "":
		_ = os.WriteFile(path, data, 0644)
	}
}

func writeCardTo(out io.Writer, data []byte) {
	encoded := base64.StdEncoding.EncodeToString(data)
	_, _ = io.WriteString(out, "\u001B]1338;")
	_, _ = io.WriteString(out, encoded)
	_, _ = io.WriteString(out, "\u001B]0m")
	_, _ = io.WriteString(out, "\n")
}
