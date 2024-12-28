package types

type PaginationData struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

type Todo struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	CreatedAt string `json:"created_at"`
	Text      string `json:"text"`
	Priority  string `json:"priority"`
	Completed bool   `json:"completed"`
	Version   int    `json:"version"`
}

type TodoResponse struct {
	PaginationData PaginationData `json:"paginationData"`
	Todos          []Todo         `json:"todos"`
}
