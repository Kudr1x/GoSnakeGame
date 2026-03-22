//go:build !js || !wasm

// Package client provides the shared game application logic and UI.
package client

func (gc *App) checkInviteLink() {
	// No invite link logic for desktop
}
