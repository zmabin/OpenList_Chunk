package handles

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/url"
	"os"
	stdpath "path"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/fs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/stream"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/cespare/xxhash/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Shared helpers (used by chunk handlers) ---

func getUserFromContext(c *gin.Context) *model.User {
	return c.Request.Context().Value(conf.UserKey).(*model.User)
}

func resolveUserPath(c *gin.Context, rawPath string) (string, error) {
	user := getUserFromContext(c)
	return user.JoinPath(rawPath)
}

func checkFileExists(ctx context.Context, path string, overwrite bool) error {
	if overwrite {
		return nil
	}
	if res, _ := fs.Get(ctx, path, &fs.GetArgs{NoLog: true}); res != nil {
		return fmt.Errorf("file exists")
	}
	return nil
}

func checkWritePermission(path string) error {
	storage, err := fs.GetStorage(path, &fs.GetStoragesArgs{})
	if err != nil {
		return err
	}
	if storage.Config().NoUpload {
		return fmt.Errorf("storage does not support upload")
	}
	return nil
}

func validateFileName(name string) error {
	if shouldIgnoreSystemFile(name) {
		return errs.IgnoredSystemFile
	}
	return nil
}

// --- Stream chunked upload (Content-Range) ---

type StreamUploadSession struct {
	pipeWriter *io.PipeWriter
	pipeReader *io.PipeReader
	totalSize  int64
	received   int64
	done       chan error
	lastActive time.Time
	mu         sync.Mutex
}

var streamUploadSessions = sync.Map{}

const streamSessionTimeout = 10 * time.Minute

func init() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			streamUploadSessions.Range(func(key, value any) bool {
				session := value.(*StreamUploadSession)
				session.mu.Lock()
				if now.Sub(session.lastActive) > streamSessionTimeout {
					session.pipeWriter.CloseWithError(fmt.Errorf("session timeout"))
					streamUploadSessions.Delete(key)
				}
				session.mu.Unlock()
				return true
			})
		}
	}()
}

func parseContentRange(header string) (start, end, total int64, err error) {
	re := regexp.MustCompile(`bytes (\d+)-(\d+)/(\d+)`)
	matches := re.FindStringSubmatch(header)
	if len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range format")
	}
	start, _ = strconv.ParseInt(matches[1], 10, 64)
	end, _ = strconv.ParseInt(matches[2], 10, 64)
	total, _ = strconv.ParseInt(matches[3], 10, 64)
	return
}

func generateStreamSessionKey(userID uint, path string, totalSize int64) string {
	return fmt.Sprintf("stream:%d:%s:%d", userID, path, totalSize)
}

func fsStreamChunked(c *gin.Context, contentRange string) {
	start, _, total, err := parseContentRange(contentRange)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	path := c.GetHeader("File-Path")
	path, err = url.PathUnescape(path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	overwrite := c.GetHeader("Overwrite") != "false"
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	path, err = user.JoinPath(path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	dir, name := stdpath.Split(path)
	if shouldIgnoreSystemFile(name) {
		common.ErrorStrResp(c, errs.IgnoredSystemFile.Error(), 403)
		return
	}

	sessionKey := generateStreamSessionKey(user.ID, path, total)

	if start == 0 {
		if !overwrite {
			if res, _ := fs.Get(c.Request.Context(), path, &fs.GetArgs{NoLog: true}); res != nil {
				common.ErrorStrResp(c, "file exists", 403)
				return
			}
		}

		pr, pw := io.Pipe()
		session := &StreamUploadSession{
			pipeWriter: pw,
			pipeReader: pr,
			totalSize:  total,
			received:   0,
			done:       make(chan error, 1),
			lastActive: time.Now(),
		}
		streamUploadSessions.Store(sessionKey, session)

		mimetype := c.GetHeader("Content-Type")
		if len(mimetype) == 0 || mimetype == "application/octet-stream" {
			mimetype = utils.GetMimeType(name)
		}

		go func() {
			s := &stream.FileStream{
				Obj: &model.Object{
					Name:     name,
					Size:     total,
					Modified: getLastModified(c),
				},
				Reader:       pr,
				Mimetype:     mimetype,
				WebPutAsTask: false,
			}
			err := fs.PutDirectly(context.Background(), dir, s, false)
			session.done <- err
		}()
	}

	sessionVal, ok := streamUploadSessions.Load(sessionKey)
	if !ok {
		common.ErrorStrResp(c, "upload session not found, please start from first chunk", 400)
		return
	}
	session := sessionVal.(*StreamUploadSession)

	session.mu.Lock()
	session.lastActive = time.Now()
	session.mu.Unlock()

	written, err := io.Copy(session.pipeWriter, c.Request.Body)
	if err != nil {
		session.pipeWriter.CloseWithError(err)
		streamUploadSessions.Delete(sessionKey)
		common.ErrorResp(c, err, 500)
		return
	}

	session.mu.Lock()
	session.received += written
	currentReceived := session.received
	session.mu.Unlock()

	if currentReceived >= total {
		session.pipeWriter.Close()
		uploadErr := <-session.done
		streamUploadSessions.Delete(sessionKey)
		if uploadErr != nil {
			common.ErrorResp(c, uploadErr, 500)
			return
		}
	}

	common.SuccessResp(c, gin.H{
		"received": currentReceived,
		"total":    total,
		"complete": currentReceived >= total,
	})
}

// --- Form chunked upload (session-based) ---

type hashVerifyingReader struct {
	io.Reader
	hasher   hash.Hash
	expected string
	verified bool
}

func (r *hashVerifyingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if n > 0 {
		r.hasher.Write(p[:n])
	}
	if err == io.EOF && !r.verified {
		actual := hex.EncodeToString(r.hasher.Sum(nil))
		if r.expected != "" && actual != r.expected {
			return n, fmt.Errorf("hash mismatch: expected %s, got %s", r.expected, actual)
		}
		r.verified = true
	}
	return n, err
}

func FsChunkInit(c *gin.Context) {
	var req struct {
		Path         string `json:"path"`
		TotalChunks  int    `json:"total_chunks"`
		LastModified int64  `json:"last_modified"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	path, err := resolveUserPath(c, req.Path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	if err := checkWritePermission(path); err != nil {
		common.ErrorStrResp(c, "no write permission", 403)
		return
	}

	uploadId := uuid.NewString()
	chunkDir := getChunkDir(uploadId)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	sessionData := map[string]interface{}{
		"user_id":       getUserFromContext(c).ID,
		"path":          req.Path,
		"total_chunks":  req.TotalChunks,
		"last_modified": req.LastModified,
		"created_at":    time.Now().Unix(),
	}

	sessionBytes, _ := json.Marshal(sessionData)
	err = os.WriteFile(stdpath.Join(chunkDir, "session.json"), sessionBytes, 0644)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, gin.H{
		"upload_id": uploadId,
	})
}

func getChunkDir(uploadId string) string {
	return stdpath.Join(conf.Conf.TempDir, "chunks", uploadId)
}

func getAndVerifyChunkSession(c *gin.Context, uploadId string) (map[string]interface{}, string, error) {
	chunkDir := getChunkDir(uploadId)
	sessionPath := stdpath.Join(chunkDir, "session.json")

	sessionBytes, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, "", fmt.Errorf("invalid upload_id or session expired")
	}

	var sessionData map[string]interface{}
	json.Unmarshal(sessionBytes, &sessionData)

	user := getUserFromContext(c)
	if float64(user.ID) != sessionData["user_id"].(float64) {
		return nil, "", fmt.Errorf("unauthorized access to chunk session")
	}

	return sessionData, chunkDir, nil
}

func buildMergeStream(chunkDir, name string, totalChunks int, lastModified time.Time, expectedHash string) (*stream.FileStream, error) {
	var readers []io.Reader
	var closers []io.Closer
	var totalSize int64

	for i := 0; i < totalChunks; i++ {
		chunkPath := stdpath.Join(chunkDir, strconv.Itoa(i))
		f, err := os.Open(chunkPath)
		if err != nil {
			for _, c := range closers {
				c.Close()
			}
			return nil, fmt.Errorf("chunk %d not found or unreadable: %w", i, err)
		}
		stat, _ := f.Stat()
		totalSize += stat.Size()

		readers = append(readers, f)
		closers = append(closers, f)
	}

	multiReader := io.MultiReader(readers...)

	hasher := xxhash.New()
	verifyingReader := &hashVerifyingReader{
		Reader:   multiReader,
		hasher:   hasher,
		expected: expectedHash,
	}

	s := &stream.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     totalSize,
			Modified: lastModified,
		},
		Reader:   verifyingReader,
		Mimetype: utils.GetMimeType(name),
	}

	s.Closers.Add(utils.CloseFunc(func() error {
		for _, c := range closers {
			c.Close()
		}
		os.RemoveAll(chunkDir)
		return nil
	}))

	return s, nil
}
