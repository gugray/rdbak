package internal

import (
	"testing"
)

type limitLengthSample struct {
	str      string
	limit    int
	expected string
}

var testLimitLengthSamples = []limitLengthSample{
	{"abc", 1, "a"},
	{"abc", 2, "ab"},
	{"abc", 3, "abc"},
	{"abc", 4, "abc"},
	{"abc.txt", 1, "a"},
	{"abc.txt", 2, "ab"},
	{"abc.txt", 3, "abc"},
	{"abc.txt", 4, "abc."},
	{"abc.txt", 5, "a.txt"},
	{"abc.txt", 6, "ab.txt"},
	{"abc.txt", 7, "abc.txt"},
	{"abc.txt", 8, "abc.txt"},
	{"ab.c.txt", 1, "a"},
	{"ab.c.txt", 2, "ab"},
	{"ab.c.txt", 3, "ab."},
	{"ab.c.txt", 4, "ab.c"},
	{"ab.c.txt", 5, "a.txt"},
	{"ab.c.txt", 6, "ab.txt"},
	{"ab.c.txt", 7, "ab..txt"},
	{"ab.c.txt", 8, "ab.c.txt"},
	{"ab.c.txt", 9, "ab.c.txt"},
}

func TestLimitLength(t *testing.T) {
	for _, sample := range testLimitLengthSamples {
		res := limitLength(sample.str, sample.limit)
		if res != sample.expected {
			t.Errorf("'%s' (%d): expected '%s' got '%s'", sample.str, sample.limit, sample.expected, res)
		}
	}
}

func TestDownloadFileManual(t *testing.T) {
	ac := newApiClient()
	ac.login("", "") // Fill in manually; don't commit :)
	// Works if CWD is rdbak/internal, i.e., where current source file is
	ac.downloadFileIfMissing(589740577, "../work")
}
