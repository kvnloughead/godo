package data

import (
	"reflect"
	"strings"

	validator "github.com/kvnloughead/godo/internal"
)

// PaginationData is a struct that contains pagination data.
type PaginationData struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func calculatePaginationData(totalRecords, page, pageSize int) PaginationData {
	if totalRecords == 0 {
		return PaginationData{}
	}

	// LastPage calculation is equivalent to ceil(totalRecords / pageSize).
	return PaginationData{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}

type Filters struct {
	Page     int
	PageSize int
	Sort     string

	// SortSafeList is a string slice containing a list of acceptable query
	// parameters for sorting.
	SortSafelist []string

	// Archive filters - default (false, false) means show only unarchived
	IncludeArchived bool
	OnlyArchived    bool

	// Completion filters
	Done   bool
	Undone bool
}

// sortColumn returns the column to sort by from the filter's Sort field.
// It panics if the sort key is not in the safelist.
func (f *Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.Trim(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection returns the direction in which the sort should occur.
// Possible return values: "ASC" and "DESC".
func (f *Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	} else {
		return "ASC"
	}
}

// limit returns the max number of items in a page, as specified by the
// `page_size` query parameter.
func (f *Filters) limit() int {
	return f.PageSize
}

// offset returns the number of rows to skip when display paginated data beyond
// page 1. Calculated from query parameters as (page - 1) * page_size.
func (f *Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validator.Validator, f Filters) {

	v.Check(f.Page >= 1, "page", "must be at least 1")
	v.Check(f.Page <= 10_000_000, "page", "must be no more than least 10,000,000")
	v.Check(f.PageSize >= 1, "page_size", "must be at least 1")
	v.Check(f.PageSize <= 100, "page_size", "must be no more than 100")

	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sorting key")

	// Validate that the boolean flags
	v.Check(reflect.TypeOf(f.IncludeArchived).Kind() == reflect.Bool, "include-archived", "must be boolean")
	v.Check(reflect.TypeOf(f.OnlyArchived).Kind() == reflect.Bool, "only-archived", "must be boolean")
	v.Check(reflect.TypeOf(f.Done).Kind() == reflect.Bool, "done", "must be boolean")
	v.Check(reflect.TypeOf(f.Undone).Kind() == reflect.Bool, "undone", "must be boolean")

	// Validate mutually exclusive flags
	if f.IncludeArchived && f.OnlyArchived {
		v.AddError("filters", "include-archived and only-archived are mutually exclusive")
	}
	if f.Done && f.Undone {
		v.AddError("filters", "done and undone are mutually exclusive")
	}
}
