package views

type baseInterface interface {
	Print()
}

type BaseView struct {
	sortModeDescending string
	sortModeAscending  string
	sortMode           string
}

func NewBaseView() *BaseView {
	return &BaseView{
		sortModeAscending:  string("ASC"),
		sortModeDescending: string("DESC"),
		sortMode:           string("ASC"),
	}
}

func (b *BaseView) sortAscending() {
	b.sortMode = b.sortModeAscending
}

func (b *BaseView) sortDescending() {
	b.sortMode = b.sortModeDescending
}
