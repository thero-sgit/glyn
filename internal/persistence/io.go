package persistence

import (
	"encoding/json"
	"time"

	"github.com/thero-sgit/glyn/internal/config"
)

func (c *ChatSession) New() *ChatSession {
	ts := time.Now().Local().UTC().Format("20060102-150405")
	c.key, _ = ProjectKey(config.Config.WorkingDir)
	c.id, _ = newSessionId(ts)
	f, _ := openSession(config.Config.WorkingDir, c.key, c.id) // temp dir
	c.file = f
	c.AppendHeaders(ts)

	return c
}

func (c *ChatSession) AppendHeaders(time string) {
	headers := map[string]string{
		"type": "header",
		"v": config.Config.Version,
		"id": c.id,
		"key": c.key,
		"created": time,
		"cmd": config.Config.WorkingDir,
		"model": config.Config.Model.Name,
	}

	title := map[string]string{
		"type": "title",
		"value" : c.Name,
	}

	if _, err := c.file.Write(append(must(json.Marshal(headers)), '\n')); err != nil {
		return
	}

	if _, err := c.file.Write(append(must(json.Marshal(title)), '\n')); err != nil {
		return
	}
}

func must(b []byte, err error) []byte {
	if err != nil {
		return []byte{}
	}

	return b
}