package repository_server

type PaginationOption[T any] struct {
	MinId T
	Limit int
}
