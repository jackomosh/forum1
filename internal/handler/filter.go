package handler

import "forum/internal/domain"

type FilterForm struct {
	CategoryID domain.CategoryID
	Kind       domain.PostFilterKind
	Search     string
	Sort       domain.SortOrder
	Page       int
}

type PaginationViewData struct {
	Page       int
	Limit      int
	Total      int
	TotalPages int
}
