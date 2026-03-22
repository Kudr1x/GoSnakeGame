//go:build js && wasm

package client

import (
	"syscall/js"
)

func (gc *App) checkInviteLink() {
	hash := js.Global().Get("location").Get("hash").String()
	if len(hash) > 1 {
		gc.roomID = hash[1:]
	}
}
