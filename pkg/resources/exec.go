package resources

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"example.com/jcfg/pkg/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Exec struct {
	Api      string
	Kind     string
	Metadata api.MetadataDef
	Spec     api.ExecSpec
}

func (e *Exec) GetApi() string {
	return e.Api
}
func (e *Exec) GetKind() string {
	return e.Kind
}
func (e *Exec) GetName() string {
	return strings.Title(e.Kind) + "::" + e.Metadata.Name
}
func (e *Exec) GetMetadata() api.MetadataDef {
	return e.Metadata
}
func (e *Exec) Fail(log *logrus.Logger, err error) error {
	if e.Spec.FailOk == true {
		log.Infof("%s: Caught err, continuing\n", e.GetName())
		log.Infof("%v\n", err)
		e.Done()
		return nil
	} else {
		e.Metadata.State.Failed = true
		return err
	}
}
func (e *Exec) Done() {
	e.Metadata.State.Completed = true
}

func (e *Exec) Init() {
	e.Metadata.State.Completed = false
	e.Metadata.State.Failed = false
}

func loadEnv(
	cmd *exec.Cmd, envSpec []api.EnvSpec, log *logrus.Logger,
) error {

	// Build output array with length of the input env spec array
	output := make([]string, len(envSpec))
	for i, env := range envSpec {
		content, err := getContent(&env.Value, log)
		if err != nil {
			return errors.Errorf(
				"Unable to fetch content for env var %s: %v", env.Name, err,
			)
		}
		output[i] = env.Name + "=" + string(content)
	}

	// Add output array to cmd as env
	cmd.Env = output

	return nil
}

func outputToString(r io.Reader, log *logrus.Logger) (string, error) {
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, r); err != nil {
		return "", errors.Errorf("Unable to copy output to string: %v", err)
	}

	return buf.String(), nil
}

func checkRunCommand(
	es *api.ExecSpec, cmd *exec.Cmd, iErr error,
	stderr *bytes.Buffer, stdout *bytes.Buffer, name string, log *logrus.Logger,
) error {

	// If command failed w/ non-exit error type
	exitErr, isExitErr := iErr.(*exec.ExitError)
	if !isExitErr && iErr != nil {
		return errors.Errorf("Command failed: %v", iErr)
	}

	// If command stder didn't match expected
	// jk we don't care
	sStderr, err := outputToString(stderr, log)
	if err != nil {
		return errors.Errorf("Reading stderr failed: %v", err)
	}
	if len(sStderr) != 0 {
		log.Infof("%s Stderr:\n%s\n", name, sStderr)
	}

	// If command stdout didn't match expected
	// jk we don't care
	sStdout, err := outputToString(stdout, log)
	if err != nil {
		return errors.Errorf("Reading stdout failed: %v", err)
	}
	if len(sStdout) != 0 {
		log.Infof("%s Stdout:\n%s\n", name, sStdout)
	}

	// If exit code doesn't match expected
	exitCode := 0
	if exitErr != nil {
		exitCode = exitErr.ExitCode()
	}
	if exitCode != es.ExitCode {
		return errors.Errorf("Exit code %d != expected %d", exitCode, es.ExitCode)
	}

	return nil
}

func (e *Exec) Apply(ictx context.Context, log *logrus.Logger) error {
	es := &e.Spec

	//	key := strings.Title(e.Kind) + "::" + e.Metadata.Name

	// Set context w/ Timeout if it exists
	var ctx context.Context
	if es.Timeout != "" {
		duration, err := time.ParseDuration(es.Timeout)
		if err != nil {
			return errors.Errorf("Unable to parse time %s: %v", es.Timeout, err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ictx, duration)
		defer cancel()
	} else {
		ctx = ictx
	}

	// Build command w/ context, path, args
	cmd := exec.CommandContext(ctx, es.Path, es.Args...)

	// Set env
	if err := loadEnv(cmd, es.Env, log); err != nil {
		return errors.Errorf("Unable to load env vars: %v", err)
	}
	// Set cwd
	cmd.Dir = es.Dir

	// Set user/group
	uid, gid, err := lookupUidGid(&es.UserID, log)
	if err != nil {
		return errors.Errorf("Unable to look up uid/gid: %v", err)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}

	// Set up output pipes
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	// Run
	err = cmd.Run()
	err = checkRunCommand(es, cmd, err, &stderr, &stdout, e.GetName(), log)
	if err != nil {
		return errors.Errorf("CheckRunCommand failed: %v", err)
	}

	// Check command
	return nil
}
