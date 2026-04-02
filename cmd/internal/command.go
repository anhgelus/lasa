package internal

type Command struct {
	Name     string
	Usage    string
	Callback func([]string)
}
