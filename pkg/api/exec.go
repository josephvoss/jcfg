package api

// Definition of exec resource
type ExecSpec struct {
	Path     string             // File path of this resource
	Args     []string           // Array of args to pass to exec
	Env      []EnvSpec          // Array of env vars to pass
	Dir      string             // CWD to set
	UserID   UserIdentifierSpec // User/group to run ass
	Timeout  string             // Time.Duration in string for exec timeout
	ExitCode int                // Expected exit code
	FailOk   bool               // Are resource failures okay? Will mark resource 'completed' instead of 'failed
}

type EnvSpec struct {
	Name  string      // Name of env var to set
	Value ContentSpec // Content of env var
}
