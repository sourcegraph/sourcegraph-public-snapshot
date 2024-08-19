package custom

const (
	NameDate                  = "Date"
	NameDateGetTimezoneOffset = "getTimezoneOffset"
)

// DateNameSection are the functions defined in the object named NameDate.
// Results here are those set to the current event object, but effectively are
// results of the host function.
var DateNameSection = map[string]*Names{
	NameDateGetTimezoneOffset: {
		Name:        NameDateGetTimezoneOffset,
		ParamNames:  []string{},
		ResultNames: []string{"tz"},
	},
}
