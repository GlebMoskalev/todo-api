package pagination

const (
	DefaultOffset = 0
	DefaultLimit  = 20
)

type Pagination struct {
	Offset int
	Limit  int
}
