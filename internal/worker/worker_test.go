/*
Copyright 2018 The Elasticshift Authors.
*/
package worker

import (
	"os"
	"testing"
)

func TestWorker(t *testing.T) {

	os.Setenv("SHIFT_HOST", "127.0.0.1")
	os.Setenv("SHIFT_PORT", "9101")
	os.Setenv("SHIFT_BUILDID", "5bcb1a970186f9aae2d5b031")
	os.Setenv("SHIFT_DIR", "/Users/ghazni/.elasticshift/storage")
	os.Setenv("WORKER_PORT", "9200")
	os.Setenv("SHIFT_TEAMID", "5a3a41f08011e098fb86b41f")
	os.Setenv("SHIFT_REPOFILE", "true")
	os.Setenv("SHIFT_LOG_LEVEL", "info")
	os.Setenv("SHIFT_LOG_FORMAT", "json")

	Run()
}
