package watcher

import (
	"encoding/json"
)

func (w *Watcher) PushMessage(data any) error {
	item, err := json.Marshal(data)
	if err != nil {
		return err
	}
	w.Broadcast <- item
	return nil
}
