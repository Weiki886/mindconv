package model

// Map is the format-independent representation of a mind map.
type Map struct {
	Root *Topic
}

// Topic is one node in a mind map topic tree.
type Topic struct {
	Title    string
	Notes    string
	Links    []Link
	Children []*Topic
}

// Link is a safe hyperlink attached to a topic.
type Link struct {
	Title string
	URL   string
}
