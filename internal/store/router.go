package store

type Router interface {
	GetAllWorkURL(rootUrl string) string
	GetWorkURL(rootUrl string) string
	UpdateWorkURL(rootUrl string) string
}

type RouterImpl struct{}

func (r *RouterImpl) GetAllWorkURL(rootUrl string) string {
	return rootUrl + "/get-all-work"
}

func (r *RouterImpl) GetWorkURL(rootUrl string) string {
	return rootUrl + "/get-work"
}

func (r *RouterImpl) UpdateWorkURL(rootUrl string) string {
	return rootUrl + "/update-work"
}
