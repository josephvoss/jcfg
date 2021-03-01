// main file for api package

// The api package contains all the data structure definitions for individal
// resources and shared API objects between separate packages.
//
package api

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Resource interface {
	GetApi() string
	GetKind() string
	GetName() string
	GetMetadata() MetadataDef
	Apply(ctx context.Context, log *logrus.Logger) error
	Fail(*logrus.Logger, error) error
	Done()
}

type MetadataDef struct {
	Name        string
	Description string
	Annotations map[string]string
	Ordering    OrderingDef
	State       StateDef
}

type OrderingDef struct {
	//	Before []string // should be pointer to resource. How? <resource>::<name>
	After []string // should be pointer to resource. How?
}

type StateDef struct {
	Completed bool
	Failed    bool
}

// Struct defining supported sources for content.
type ContentSpec struct {
	Type        string // Name of source to use. Used as key in this struct.
	String      string // Source file content from this string.
	LocalSource string // Source file content from file with this path on host.
	Secret      SecretSpec
	HTTPSource  HTTPSourceSpec
}

type SecretSpec struct {
}

// Empty stub for now.
type HTTPSourceSpec struct {
}

type UserIdentifierSpec struct {
	User  string // Named owner user. Will be passed to user.Lookup()
	Uid   uint32 // Numeric User ID. Either set this or Owner
	Group string // Named owner group. Will be passed to user.LookupGroup()
	Gid   uint32 // Number Group ID. Either set this or Group
}
