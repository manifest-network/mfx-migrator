package state

type Status int

const (
	CREATED Status = iota + 1
	CLAIMED
	PROCESSING
	FAILED
)

func (s Status) String() string {
	return [...]string{"created", "claimed", "processing", "failed"}[s-1]
}

// EnumIndex returns the enum index of a LocalStatus.
func (s Status) EnumIndex() int {
	return int(s)
}
