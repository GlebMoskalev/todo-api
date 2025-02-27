package status

type Status string

const (
	Planned    Status = "planned"
	InProgress Status = "in_progress"
	Completed  Status = "completed"
	Canceled   Status = "canceled"
)

func IsValidStatus(s Status) bool {
	switch s {
	case Planned, InProgress, Completed, Canceled:
		return true
	default:
		return false
	}
}
