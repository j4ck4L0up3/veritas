package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"
	"uuid"

	"github.com/j4ck4L0up3/veritas/db"
	"github.com/j4ck4L0up3/veritas/internal/logger"
)

const supportedDigestHash = "sha256"

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

type appendUploadBytesParams struct {
	uploadPath    string
	id            uuid.UUID
	repo          string
	bytesReceived uint64
	body          io.ReadCloser
	ctx           context.Context
	lgr           logger.Logger
	qry           *db.Queries
}

// appends PATCH/PUT upload bytes to the end of the upload path
// logs errors and returns range_offset and ok flag
func appendUploadBytes(p appendUploadBytesParams) (uint64, bool) {
	f, err := os.OpenFile(
		p.uploadPath,
		os.O_APPEND|io.SeekEnd,
		0o600,
	)
	if err != nil {
		p.lgr.Errorf("unable to open upload file: %v", err)
		return 0, false
	}

	defer func() {
		if err := f.Close(); err != nil {
			p.lgr.Warnf("closing upload file failed: %v", err)
		}
	}()

	written, err := io.Copy(f, p.body)
	if err != nil {
		p.lgr.Errorf("unable to copy bytes from request body to upload file: %v", err)
		return 0, false
	}

	if err = f.Sync(); err != nil {
		p.lgr.Warn(
			"failed to sync upload file to disk; may crash on system off/crash",
			"error",
			err,
		)
	}

	params := db.UpdateUploadSessionBytesReceivedParams{
		Uuid:          p.id,
		Repo:          p.repo,
		BytesReceived: p.bytesReceived + uint64(written),
	}

	err = p.qry.UpdateUploadSessionBytesReceived(p.ctx, params)
	if err != nil {
		p.lgr.Errorf("unable to update bytes received: %v", err)

		err := f.Truncate(int64(p.bytesReceived))
		if err != nil {
			p.lgr.Errorf(
				"unable to truncate file to the bytes successfully received: %v",
				err,
			)
		}

		return 0, false
	}

	rangeOffset := p.bytesReceived + uint64(written) - 1

	return rangeOffset, true
}

// checks the validity of the digest by comparing against stored upload bytes
// logs errors and returns readonly upload file and ok flag
func isValidDigest(uploadPath, digest string, lgr logger.Logger) (error, bool) {
	digestParts := strings.Split(digest, ":")
	if len(digestParts) != 2 {
		lgr.Warnf("client digest is unsupported: %s", digest)
		return nil, false
	}

	hashType := digestParts[0]
	givenDigest := digestParts[1]

	if hashType != supportedDigestHash {
		lgr.Warnf("client digest hash type is unsupported: %s", hashType)
		return nil, false
	}

	f, err := os.Open(uploadPath)
	if err != nil {
		lgr.Errorf("unable to open upload file: &v", err)
		return err, false
	}

	defer func() {
		if err := f.Close(); err != nil {
			lgr.Warnf("closing upload file failed: %v", err)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		lgr.Errorf("unable to compute hash for uploaded bytes: %v", err)
		return err, false
	}

	expectedDigest := hex.EncodeToString(h.Sum(nil))
	if givenDigest != expectedDigest {
		return nil, false
	}

	return nil, true
}
