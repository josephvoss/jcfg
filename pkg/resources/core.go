package resources

import (
	"io/ioutil"
	"os/user"
	"strconv"
	"strings"

	"example.com/jcfg/pkg/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Retrieve the content source for a resource
func getContent(con *api.ContentSpec, log *logrus.Logger) (
	[]byte, error,
) {
	switch strings.ToLower(con.Type) {
	case "string":
		return []byte(con.String), nil
	case "localsource":
		sourceData, err := ioutil.ReadFile(con.LocalSource)
		if err != nil {
			return nil, errors.Errorf(
				"Unable to read LocalSource file %s: %v", con.LocalSource, err,
			)
		}
		return sourceData, nil
	case "httpsource":
		return nil, errors.Errorf("HTTPSource currently not implemented\n")
	case "secret":
		return nil, errors.Errorf("Secret currently not implemented\n")
	default:
		return nil, errors.Errorf("Unknown content type %s\n", con.Type)
	}
	return nil, nil
}

// Return UID and GID from UserIdentifierSpec. Just drops int values if set,
// otherwise does lookups for user/group name
func lookupUidGid(
	uis *api.UserIdentifierSpec,
	log *logrus.Logger,
) (
	uint32, uint32, error,
) {
	// Look up uid and gid for resource def if string names not set
	var expectedUid uint32
	if uis.User == "" {
		expectedUid = uis.Uid
	} else {
		userDef, err := user.Lookup(uis.User)
		if err != nil {
			return 0, 0, errors.Errorf("Unable to lookup user %s: %v", uis.User, err)
		}
		iUid, err := strconv.Atoi(userDef.Uid)
		if err != nil {
			return 0, 0, errors.Errorf(
				"Unable to convert str user %s to int: %v", userDef.Uid, err,
			)
		}
		log.Debugf("UID is %v\n", iUid)
		expectedUid = uint32(iUid)
	}
	var expectedGid uint32
	if uis.Group == "" {
		expectedGid = uis.Gid
	} else {
		groupDef, err := user.LookupGroup(uis.Group)
		if err != nil {
			return 0, 0, errors.Errorf("Unable to lookup group %s: %v", uis.Group, err)
		}
		iGid, err := strconv.Atoi(groupDef.Gid)
		if err != nil {
			return 0, 0, errors.Errorf(
				"Unable to convert str group %s to int: %v", groupDef.Gid, err,
			)
		}
		expectedGid = uint32(iGid)
	}
	return expectedUid, expectedGid, nil
}
