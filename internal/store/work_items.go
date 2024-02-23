package store

type Meta struct {
	TotalItems   int `json:"totalItems"`
	ItemCount    int `json:"itemCount"`
	ItemsPerPage int `json:"itemsPerPage"`
	TotalPages   int `json:"totalPages"`
	CurrentPage  int `json:"currentPage"`
}

type WorkItems struct {
	Items []WorkItem `json:"items"`
	Meta  Meta       `json:"meta"`
}
