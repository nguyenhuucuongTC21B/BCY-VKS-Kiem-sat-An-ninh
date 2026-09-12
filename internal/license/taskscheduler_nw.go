//go:build !windows

package license

// ScheduledTask placeholder cho non-Windows
type ScheduledTask struct {
	Name        string
	Status      string
	Command     string
	Suspicious  bool
	Reason      string
}

func scanScheduledTasks() []ScheduledTask { return nil }
