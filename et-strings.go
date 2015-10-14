//
// Handles Wolfenstein: Enemy Territory related string parsing
package main
import "strings"

//
// Strips ET color codes (e.g. ^7) from string and
// returns a clean string
func StripColors(text string) (string) {
	var buf string
	prevWasCaret := false

	for _, c := range text {

		if c == '^' {
			if prevWasCaret {
				buf = buf + string(c)
			} else {
				prevWasCaret = true
			}
		} else {
			if !prevWasCaret {
				buf = buf + string(c)
			}
			prevWasCaret = false
		}
	}

	return buf
}

//
// A status response from an ET server
type StatusResponse struct {
	Keys    map[string]string
	Players []string
}

//
// Parses a received server status response
func ParseStatusResponse(statusResponse string) (sr *StatusResponse) {
	sr = &StatusResponse{}
	rows := strings.Split(statusResponse, "\n")
	sr.Keys = make(map[string]string)

	for _, player := range rows[2:len(rows) - 1] {
		sr.Players = append(sr.Players, strings.Split(player, "\"")[1])
	}

	keys := strings.Split(rows[1], "\\")
	key := ""

	for _, value := range keys {
		if len(key) == 0 {
			key = value
		} else {
			sr.Keys[key] = value
			key = ""
		}
	}

	return sr
}
