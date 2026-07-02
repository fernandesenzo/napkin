package handler

type saveNapkinRequest struct {
	Code    string `json:"code"`
	Content string `json:"content"`
}

type saveNapkinResponse struct {
	Content string `json:"content"`
}

type getNapkinResponse struct {
	Code    string `json:"code"`
	Content string `json:"content"`
}
