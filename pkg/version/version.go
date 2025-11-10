package version

var (
	Tag     = "dev"
	Commit  = ""
	BuiltAt = ""
)

func String() string { return Tag }
