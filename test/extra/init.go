package extra

// Init initializes the HTML and sets the subtitle.
func (o *OptionalSpam) Init(title string, spam bool) {
	o.askewInit(title, spam)
	o.Subtitle.Set("subtitle spam")
}
