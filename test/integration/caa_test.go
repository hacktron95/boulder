// +build integration

package integration

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/letsencrypt/boulder/test"
)

func TestCAALogChecker(t *testing.T) {
	t.Parallel()

	os.Setenv("DIRECTORY", "http://boulder:4001/directory")
	c, err := makeClient()
	test.AssertNotError(t, err, "makeClient failed")

	domains := []string{random_domain()}
	result, err := authAndIssue(c, nil, domains)
	test.AssertNotError(t, err, "authAndIssue failed")
	test.AssertEquals(t, result.Order.Status, "valid")
	test.AssertEquals(t, len(result.Order.Authorizations), 1)

	// Should be no specific output, since everything is good
	cmd := exec.Command("bin/caa-log-checker", "-ra-log", "/var/log/boulder-ra.log", "-va-logs", "/var/log/boulder-va.log")
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	out, err := cmd.Output()
	test.AssertNotError(t, err, "caa-log-checker failed")
	test.AssertEquals(t, string(out), "")
	test.AssertEquals(t, stdErr.String(), "")

	// Should be output, issuances in boulder-ra.log won't match an empty
	// va log. Because we can't control what happens before this test
	// we don't know how many issuances there have been. We just
	// test for caa-log-checker outputting _something_ since any
	// output, with a 0 exit code, indicates it's found bad issuances.
	tmp, err := ioutil.TempFile(os.TempDir(), "boulder-va-empty")
	test.AssertNotError(t, err, "failed to create temporary file")
	defer os.Remove(tmp.Name())
	cmd = exec.Command("bin/caa-log-checker", "-ra-log", "/var/log/boulder-ra.log", "-va-logs", tmp.Name())
	stdErr.Reset()
	cmd.Stderr = &stdErr
	out, err = cmd.Output()
	test.AssertError(t, err, "caa-log-checker didn't")

	test.AssertEquals(t, string(out), "")
	test.AssertNotEquals(t, stdErr.String(), "")
}
