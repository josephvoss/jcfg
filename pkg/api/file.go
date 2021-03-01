package api

// Definition of file resource
type FileSpec struct {
	Ensure  string // Supported types: present, absent, directory, or link
	Path    string // File path of this resource
	UserID  UserIdentifierSpec
	Mode    string      // String of 4 digit unix permissions mode
	Target  string      // If ensure is link, set link target
	Content ContentSpec // Struct to define content
}
