package store_test

type MockRouterImpl struct{}

func (r *MockRouterImpl) GetAllWorkURL(rootUrl string) string {
	return rootUrl + "/get-all-work"
}

func (r *MockRouterImpl) GetWorkURL(rootUrl string) string {
	return rootUrl + "/get-work"
}

func (r *MockRouterImpl) UpdateWorkURL(rootUrl string) string {
	return rootUrl + "/update-work-failure"
}
