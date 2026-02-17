package request

type PaginationRequest struct {
	Page     int `json:"page" query:"page" validate:"min=1"`
	PageSize int `json:"page_size" query:"page_size" validate:"min=1,max=100"`
}

func GetDefaultPagination() PaginationRequest {
	return PaginationRequest{
		Page:     1,
		PageSize: 20,
	}
}

func (p PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

func (p PaginationRequest) GetLimit() int {
	return p.PageSize
}

type SortRequest struct {
	SortBy    string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=created_at updated_at name status"`
	SortOrder string `json:"sort_order" query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

func GetDefaultSort() SortRequest {
	return SortRequest{
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

type FilterRequest struct {
	Name     *string `json:"name,omitempty" query:"name"`
	Status   *string `json:"status,omitempty" query:"status" validate:"omitempty,oneof=active inactive archived"`
	FromDate *string `json:"from_date,omitempty" query:"from_date"`
	ToDate   *string `json:"to_date,omitempty" query:"to_date"`
}

type ListUsersRequest struct {
	PaginationRequest
	SortRequest
	FilterRequest
}
