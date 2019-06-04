package tests


type IndexTest struct {
	s   string
	sep string
	out int
}

var lastIndexTests = []IndexTest{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
}