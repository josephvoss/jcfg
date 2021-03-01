package api

// Definition of exec resource
type ExecSpec struct {
	Path     string    // File path of this resource
	Args     []string  // File path of this resource
	Env      []EnvSpec // File path of this resource
	Dir      string    // File path of this resource
	UserID   UserIdentifierSpec
	Timeout  string
	ExitCode int
	FailOk   bool
}

type EnvSpec struct {
	Name  string
	Value ContentSpec
}
