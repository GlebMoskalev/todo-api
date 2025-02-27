package priority

type Priority string

const (
	Low    Priority = "low"
	Medium Priority = "medium"
	High   Priority = "high"
	Urgent Priority = "urgent"
)

func IsValidPriority(p Priority) bool {
	switch p {
	case Low, Medium, High, Urgent:
		return true
	default:
		return false
	}
}
