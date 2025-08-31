package db_dao

// GetEndPoint Get选择器
type GetEndPoint[T any] struct {
	Model      *T
	Table      string
	Conditions map[string]any
	Appends    []string
	Options    []string
	Fields     []string
}

// SelectEndPoint Select选择器
type SelectEndPoint[T any] struct {
	Model      *[]T
	Table      string
	Conditions map[string]any
	Appends    []string
	Options    []string
	Fields     []string
}

// PageEndPoint Select分页选择器
type PageEndPoint[T any] struct {
	Model      *[]T
	Table      string
	Conditions map[string]any
	SortField  string // SortField 用于指定排序字段
	SortOrder  string // SortOrder 用于指定排序顺序 (ASC/DESC)
	PageNo     int32
	PageSize   int32
	Fields     []string
}

// UpdateEndPoint Update选择器
type UpdateEndPoint[T any] struct {
	Table      string
	Rows       map[string]any
	Conditions map[string]any
	Appends    []string
	Options    []string
}

// InsertEndpoint Insert选择器
type InsertEndpoint[T any] struct {
	Table string
	Rows  map[string]any
}

// BatchInsertEndpoint BatchInsert选择器
type BatchInsertEndpoint[T any] struct {
	Table string
	Rows  []map[string]any
}

// DeleteEndPoint Delete选择器
type DeleteEndPoint[T any] struct {
	Table      string
	Conditions map[string]any
}
