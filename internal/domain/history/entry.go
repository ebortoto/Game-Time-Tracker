package history

import "time"

// Entry represents the persisted aggregate playtime for a game.
type Entry struct {
	GameName       string
	TotalTimeSecs  int64
	LastPlayedDate time.Time
}
