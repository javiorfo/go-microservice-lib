package pagination

import (
	"errors"
	"strconv"

	"github.com/javiorfo/go-microservice-lib/response"
)

type Page struct {
	Page      int
	Size      int
	SortBy    string
	SortOrder string
}

func DefaultPage() Page {
	return Page{
		Page:      1,
		Size:      10,
		SortBy:    "id",
		SortOrder: "asc",
	}
}

func NewPage(page, size, sortBy, sortOrder string) (*Page, error) {
	p := Page{SortBy: sortBy}

	if pageInt, err := strconv.Atoi(page); err != nil {
		return nil, errors.New("'page' parameter must be a number")
	} else {
		p.Page = pageInt
	}

	if sizeInt, err := strconv.Atoi(size); err != nil {
		return nil, errors.New("'size' parameter must be a number")
	} else {
		p.Size = sizeInt
	}

	if sortOrder == "asc" || sortOrder == "desc" {
		p.SortOrder = sortOrder
		return &p, nil
	} else {
		return nil, errors.New("'sortOrder' parameter must be 'asc' or 'desc'")
	}
}

func Paginator(p Page, total int) response.Pagination {
	return response.Pagination{
		PageNumber: p.Page,
		PageSize:   p.Size,
		Total:      total,
	}
}
