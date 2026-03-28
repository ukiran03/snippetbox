package models

// TemplateData will be the data used while rendering the template
type TemplateData struct {
	CurrentYear     int
	Snippet         Snippet
	Snippets        []Snippet
	Form            interface{}
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}
