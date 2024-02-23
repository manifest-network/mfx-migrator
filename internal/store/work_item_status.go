package store

type WorkItemStatus int

const (
	CREATED WorkItemStatus = iota + 1
	CLAIMED
	MIGRATING
	COMPLETED
	FAILED
)

func (s WorkItemStatus) String() string {
	return [...]string{"created", "claimed", "migrating", "completed", "failed"}[s-1]
}

// EnumIndex returns the enum index of a LocalWorkItemStatus.
func (s WorkItemStatus) EnumIndex() int {
	return int(s)
}
