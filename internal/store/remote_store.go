package store

// RemoteStore is a store that interacts with a remote database.
// It uses a Router to determine the URLs for the various operations.
type RemoteStore struct {
	Router Router
}

// NewRemoteStore creates a new RemoteStore.
// The Router is used to determine the URLs for the various operations.
func NewRemoteStore(router Router) *RemoteStore {
	return &RemoteStore{Router: router}
}
