package status

type Status string

const (
	Planned    Status = "planned"
	InProgress Status = "in_progress"
	Completed  Status = "completed"
	Canceled   Status = "canceled"
)
