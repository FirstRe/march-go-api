package repositories

type WhereArgs struct {
	Where     interface{}
	WhereArgs []interface{}
}

type FindParams struct {
	WhereArgs   []WhereArgs
	Preload     []string
	SelectField []string
	OrderBy     string
	Limit       *int
	Offset      *int
}
