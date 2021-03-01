package resources

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"

	"example.com/jcfg/pkg/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type File struct {
	Api      string
	Kind     string
	Metadata api.MetadataDef
	Spec     api.FileSpec
}

func (f *File) GetApi() string {
	return f.Api
}
func (f *File) GetKind() string {
	return f.Kind
}
func (f *File) GetName() string {
	return strings.Title(f.Kind) + "::" + f.Metadata.Name
}
func (f *File) GetMetadata() api.MetadataDef {
	return f.Metadata
}
func (f *File) Fail(log *logrus.Logger, err error) error {
	f.Metadata.State.Failed = true
	return err
}
func (f *File) Done() {
	f.Metadata.State.Completed = true
}

func (f *File) Init() {
	f.Metadata.State.Completed = false
	f.Metadata.State.Failed = false
}

// Get current permissions of file. Pass in filepath via string, outputs unix
// perm string, error
func getFileMode(fp string, log *logrus.Logger) (os.FileMode, error) {
	log.Debugf("Running lstat %s\n", fp)
	fi, err := os.Lstat(fp)
	if err != nil {
		return 0, errors.Errorf(
			"Unable to lstat %s for permissions: %v", fp, err,
		)
	}
	return fi.Mode(), nil
}

// Ensure mode/permissions of file is set to expected
func ensureMode(f *api.FileSpec, log *logrus.Logger) error {
	// Check
	expectedMode := "0600"
	// Should do stricter type checking here for valid modes, but w/e
	if f.Mode != "" {
		expectedMode = f.Mode
	}
	// Octal 32 bit number
	intMode, err := strconv.ParseInt(expectedMode, 8, 32)
	if err != nil {
		return errors.Errorf(
			"Unable to convert string mode of %s to octal int32 %s: %v", f.Path,
			expectedMode, err,
		)
	}
	setMode := os.FileMode(intMode)

	fm, err := getFileMode(f.Path, log)
	if err != nil {
		return errors.Errorf(
			"Unable to check initial mode for %s: %v", f.Path, err,
		)
	}
	if setMode == fm {
		log.Debugf("Mode of %s matches expected\n", f.Path)
		return nil
	}

	// Set
	log.Debugf("Setting mode of %s to %v\n", f.Path, setMode)
	if err := os.Chmod(f.Path, setMode); err != nil {
		return errors.Errorf(
			"Unable to set mode of %s to %s: %v", f.Path, expectedMode, err,
		)
	}

	// Check again
	fm, err = getFileMode(f.Path, log)
	if err != nil {
		return errors.Errorf(
			"Unable to check final mode for %s: %v", f.Path, err,
		)
	}
	if setMode.Perm() != fm.Perm() {
		return errors.Errorf(
			"Unable to persistently set mode of %s from %v to %v", f.Path, fm, setMode)
	}
	return nil
}

// Ensure directory exists. We're going to manage permissions later, so set
// 0700 for now
func ensureDirPresent(f *api.FileSpec, log *logrus.Logger) error {
	// Lstat to read file info
	_, err := os.Lstat(f.Path)
	// If we get error for path doesn't exist
	if os.IsNotExist(err) {
		// create directory w/ dummy perms
		if err := os.MkdirAll(f.Path, os.FileMode(0700)); err != nil {
			return errors.Errorf(
				"Unable to mkdir to ensure directory %s: %v", f.Path, err,
			)
		}
		return nil
		// If we got a seperate error, throw
	} else if err != nil {
		return errors.Errorf(
			"Unable to lstat to ensure directory %s: %v", f.Path, err,
		)
	}

	return nil
}

func ensureFilePresent(f *api.FileSpec, log *logrus.Logger) error {
	// Lstat to see if file exists
	_, err := os.Lstat(f.Path)
	// If no error - we're good
	if err == nil {
		return nil
	}
	// If error and not IsNotExist, throw
	if !os.IsNotExist(err) {
		return errors.Errorf(
			"Unable to lstat %s to ensure present: %v", f.Path, err,
		)

	}
	// If isNotExist error, create file
	if _, err = os.Create(f.Path); err != nil {
		return errors.Errorf(
			"Unable to create %s to ensure present: %v", f.Path, err,
		)
	}
	return nil
}

// Ensure file is absent. Run os.RemoveAll no matter current state (returns nil
// if path doesn't exist, so we don't need to check ourselves).
func ensureAbsent(f *api.FileSpec, log *logrus.Logger) error {
	if err := os.RemoveAll(f.Path); err != nil {
		return errors.Errorf(
			"Unable to remove %s to ensure absent: %v", f.Path, err,
		)
	}
	return nil
}

// Get current ownership of file. Pass in filepath via string, outputs uid,
// gid, error
func getFileOwnership(fp string) (uint32, uint32, error) {
	fi, err := os.Lstat(fp)
	if err != nil {
		return 0, 0, errors.Errorf(
			"Unable to lstat %s for ownership: %v", fp, err,
		)
	}
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, errors.Errorf(
			"Unable to cast stat of %s to syscall struct: %v", fp, err,
		)
	}
	return stat.Uid, stat.Gid, nil
}

// Check current ownership, ensure matches expected
// Looks up string of owner and group params.
func ensureOwners(f *api.FileSpec, log *logrus.Logger) error {

	expectedUid, expectedGid, err := lookupUidGid(&f.UserID, log)
	if err != nil {
		return errors.Errorf("Unable to look up uid/gid: %v", err)
	}

	// Get current ownership
	uid, gid, err := getFileOwnership(f.Path)
	if err != nil {
		return errors.Errorf(
			"Unable to get initial ownership of %s: %v", f.Path, err,
		)
	}

	// Check uid and gid
	if expectedUid == uid && expectedGid == gid {
		log.Debugf("Ownership of %s matches expected.\n", f.Path)
		return nil
	}

	// Change uid and gid
	log.Debugf(
		"Setting uid from %d to %d and gid from %d to %d", uid, expectedUid, gid,
		expectedGid,
	)
	if err = os.Chown(f.Path, int(expectedUid), int(expectedGid)); err != nil {
		return errors.Errorf(
			"Unable to chown %s to uid %v and gid %v: %v", f.Path, expectedUid,
			expectedGid, err,
		)
	}

	// Verify uid and gid are now what we expect.
	// Check uid and gid
	uid, gid, err = getFileOwnership(f.Path)
	if err != nil {
		return errors.Errorf("Unable to get final ownership of %s: %v",
			f.Path, err,
		)
	}
	if expectedUid != uid || expectedGid != gid {
		return errors.Errorf(
			"Unable to persistently set ownership of %s to uid %d and gid %d",
			f.Path, expectedUid, expectedGid,
		)
	}
	return nil
}

// Read a file, compare content to input byte array. If diff return false
func checkFileContent(filepath string, expectedData []byte) (bool, error) {
	actualData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return false, errors.Errorf("Unable to open file %s: %v", filepath, err)
	}
	return bytes.Compare(actualData, expectedData) == 0, nil
}

func ensureFileContent(f *api.FileSpec, log *logrus.Logger) error {
	// Get file content
	log.Debugf("Loading expected content of %s\n", f.Path)
	expectedContent, err := getContent(&f.Content, log)
	log.Debugf("expected content of %s is %s\n", f.Path, string(expectedContent))
	if err != nil {
		return errors.Errorf(
			"Unable to load expected file content for %s: %v", f.Path, err,
		)
	}

	// Check current content
	log.Debugf("Checking current content of %s\n", f.Path)
	fileDiff, err := checkFileContent(f.Path, expectedContent)
	if err != nil {
		return errors.Errorf("Unable to check file content %s: %v", f.Path, err)
	}
	if fileDiff == true {
		log.Debugf("File content matches expected\n")
		return nil
	}

	// Set content (we're guaranteed the file exists by this point, and perms on
	// existing files are ignored)
	log.Debugf("Writing content to %s\n", f.Path)
	if err = ioutil.WriteFile(f.Path, expectedContent, 000); err != nil {
		return errors.Errorf("Unable to set file content %s: %v", f.Path, err)
	}

	// Verify file content was set correctly
	fileDiff, err = checkFileContent(f.Path, expectedContent)
	if err != nil {
		return errors.Errorf("Unable to check file content %s: %v", f.Path, err)
	}
	if fileDiff != true {
		return errors.Errorf(
			"Unable to persistently set content of %s: %v", f.Path, nil,
		)
	}
	return nil
}

// Readlink filePath, return if matches target
func checkLink(filePath string, target string) (bool, error) {
	actual, err := os.Readlink(filePath)
	if err != nil {
		return false, errors.Errorf("Unable to read link %s: %v", filePath, err)
	}
	// If already set, do nothing
	return actual == target, nil
}
func ensureLink(f *api.FileSpec, log *logrus.Logger) error {
	// check. May fail - what happens when link doesn't exist?
	// Func will ever only return fs.patherrors
	matches, err := checkLink(f.Path, f.Target)
	if err != nil {
		return errors.Errorf("Unable to check initial link %s: %v", f.Path, err)
	}
	// If already set, do nothing
	if matches == true {
		log.Debugf("Link target matches\n")
		return nil
	}
	// set
	if err = os.Symlink(f.Path, f.Target); err != nil {
		return errors.Errorf("Unable to create link %s: %v", f.Path, err)
	}
	// check
	matches, err = checkLink(f.Path, f.Target)
	if err != nil {
		return errors.Errorf(
			"Unable to check link after ensure %s: %v", f.Path, err,
		)
	}
	// If not set, bomb
	if matches == false {
		return errors.Errorf(
			"Unable to persistently set target of %s: %v", f.Path, nil,
		)
	}
	return nil
}

func (f *File) Apply(ctx context.Context, log *logrus.Logger) error {
	// Key off ensure setting
	fs := &f.Spec
	switch fs.Ensure {
	case "absent":
		log.Debugf("Ensuring %s is absent.\n", fs.Path)
		return ensureAbsent(fs, log)
	case "directory":
		log.Debugf("Ensuring directory\n")
		// Ensure dir exists. If fails, throw
		if err := ensureDirPresent(fs, log); err != nil {
			return err
		}
		// Ensure ownership is correct. If fails, throw
		if err := ensureOwners(fs, log); err != nil {
			return err
		}
		// Ensure permissions are correct. If fails, throw
		if err := ensureMode(fs, log); err != nil {
			return err
		}
	case "present":
		log.Debugf("Ensuring present\n")
		// Ensure file exists. If fails, throw
		if err := ensureFilePresent(fs, log); err != nil {
			return err
		}
		// Ensure ownership is correct. If fails, throw
		if err := ensureOwners(fs, log); err != nil {
			return err
		}
		// Ensure permissions are correct. If fails, throw
		if err := ensureMode(fs, log); err != nil {
			return err
		}
		// Ensure content is correct. If fails, throw
		if err := ensureFileContent(fs, log); err != nil {
			return err
		}
	case "link":
		log.Debugf("Ensuring link\n")
		// Ensure link is correct/exists. If fails, throw
		if err := ensureLink(fs, log); err != nil {
			return nil
		}
		// Ensure ownership is correct. If fails, throw
		if err := ensureOwners(fs, log); err != nil {
			return err
		}
	default:
		return errors.Errorf("Cannot ensure %s", fs.Ensure)
	}
	return nil
}
