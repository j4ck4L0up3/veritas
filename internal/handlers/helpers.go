package handlers

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/j4ck4L0up3/veritas/internal/logger"
)

// checks that Content-Range header is valid
// logs errors and returns false if not
func isValidContentRange(contentRange string, bytesReceived uint64, lgr logger.Logger) bool {
	start, err := strconv.Atoi(strings.Split(contentRange, "-")[0])
	if err != nil {
		lgr.Debugf("Content-Range from request: %s", contentRange)
		return false
	}

	if uint64(start) != bytesReceived {
		lgr.Debug(
			"offset received does not equal stored bytes_received",
			"offset",
			start,
			"bytes_received",
			bytesReceived,
		)
		return false
	}

	return true
}

// appends PATCH/PUT upload bytes to the end of the upload path
// logs errors and returns bytes written and ok flag
func appendUploadBytes(
	uploadPath string,
	body io.ReadCloser,
	lgr logger.Logger,
) (uint64, *os.File, bool) {
	f, err := os.OpenFile(
		uploadPath,
		os.O_APPEND|io.SeekEnd,
		0o600,
	)
	if err != nil {
		lgr.Errorf("unable to open upload file: %v", err)
		return 0, nil, false
	}

	written, err := io.Copy(f, body)
	if err != nil {
		lgr.Errorf("unable to copy bytes from request body to upload file: %v", err)
		return 0, nil, false
	}

	if err = f.Sync(); err != nil {
		lgr.Warn(
			"failed to sync upload file to disk; may crash on system off/crash",
			"error",
			err,
		)
	}

	return uint64(written), f, true
}
