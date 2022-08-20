package settings

type settings struct {
	Logging  bool
	Setting2 bool
	Setting3 bool
}

var Values = settings{
	Logging:  false,
	Setting2: false,
	Setting3: false,
}
