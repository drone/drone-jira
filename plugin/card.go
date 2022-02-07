package plugin

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
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
		ioutil.WriteFile(path, data, 0644)
	}
}

func writeCardTo(out io.Writer, data []byte) {
	encoded := base64.StdEncoding.EncodeToString(data)
	io.WriteString(out, "\u001B]1338;")
	io.WriteString(out, encoded)
	io.WriteString(out, "\u001B]0m")
	io.WriteString(out, "\n")
}
