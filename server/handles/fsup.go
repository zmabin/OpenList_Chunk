package handles

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"net/url"
	"os"
	stdpath "path"
	"strconv"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/fs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/internal/stream"
	"github.com/OpenListTeam/OpenList/v4/internal/task"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

func getLastModified(c *gin.Context) time.Time {
	now := time.Now()
	lastModifiedStr := c.GetHeader("Last-Modified")
	lastModifiedMillisecond, err := strconv.ParseInt(lastModifiedStr, 10, 64)
	if err != nil {
		return now
	}
	lastModified := time.UnixMilli(lastModifiedMillisecond)
	return lastModified
}

func shouldIgnoreSystemFile(filename string) bool {
	if setting.GetBool(conf.IgnoreSystemFiles) {
		return utils.IsSystemFile(filename)
	}
	return false
}

func FsStream(c *gin.Context) {
	defer func() {
		if n, _ := io.ReadFull(c.Request.Body, []byte{0}); n == 1 {
			_, _ = utils.CopyWithBuffer(io.Discard, c.Request.Body)
		}
		_ = c.Request.Body.Close()
	}()

	// Chunk dispatch: if Content-Range header is present, use chunked stream upload
	contentRange := c.GetHeader("Content-Range")
	if contentRange != "" {
		fsStreamChunked(c, contentRange)
		return
	}

	fsStreamDirect(c)
}

// fsStreamDirect handles direct (non-chunked) stream upload — upstream FsStream logic
func fsStreamDirect(c *gin.Context) {
	path := c.GetHeader("File-Path")
	path, err := url.PathUnescape(path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	asTask := c.GetHeader("As-Task") == "true"
	overwrite := c.GetHeader("Overwrite") != "false"
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	path, err = user.JoinPath(path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	if !overwrite {
		if res, _ := fs.Get(c.Request.Context(), path, &fs.GetArgs{NoLog: true}); res != nil {
			common.ErrorStrResp(c, "file exists", 403)
			return
		}
	}
	dir, name := stdpath.Split(path)
	if shouldIgnoreSystemFile(name) {
		common.ErrorStrResp(c, errs.IgnoredSystemFile.Error(), 403)
		return
	}
	size := c.Request.ContentLength
	if size < 0 {
		sizeStr := c.GetHeader("X-File-Size")
		if sizeStr != "" {
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil {
				common.ErrorResp(c, err, 400)
				return
			}
		}
	}
	h := make(map[*utils.HashType]string)
	if md5 := c.GetHeader("X-File-Md5"); md5 != "" {
		h[utils.MD5] = md5
	}
	if sha1 := c.GetHeader("X-File-Sha1"); sha1 != "" {
		h[utils.SHA1] = sha1
	}
	if sha256 := c.GetHeader("X-File-Sha256"); sha256 != "" {
		h[utils.SHA256] = sha256
	}
	mimetype := c.GetHeader("Content-Type")
	if len(mimetype) == 0 {
		mimetype = utils.GetMimeType(name)
	}
	s := &stream.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     size,
			Modified: getLastModified(c),
			HashInfo: utils.NewHashInfoByMap(h),
		},
		Reader:       c.Request.Body,
		Mimetype:     mimetype,
		WebPutAsTask: asTask,
	}
	var t task.TaskExtensionInfo
	if asTask {
		t, err = fs.PutAsTask(c.Request.Context(), dir, s)
	} else {
		err = fs.PutDirectly(c.Request.Context(), dir, s)
	}
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if t == nil {
		common.SuccessResp(c)
		return
	}
	common.SuccessResp(c, gin.H{
		"task": getTaskInfo(t),
	})
}

func FsForm(c *gin.Context) {
	defer func() {
		if n, _ := io.ReadFull(c.Request.Body, []byte{0}); n == 1 {
			_, _ = utils.CopyWithBuffer(io.Discard, c.Request.Body)
		}
		_ = c.Request.Body.Close()
	}()
	path := c.GetHeader("File-Path")
	path, err := url.PathUnescape(path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	asTask := c.GetHeader("As-Task") == "true"
	overwrite := c.GetHeader("Overwrite") != "false"
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	path, err = user.JoinPath(path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	if !overwrite {
		if res, _ := fs.Get(c.Request.Context(), path, &fs.GetArgs{NoLog: true}); res != nil {
			common.ErrorStrResp(c, "file exists", 403)
			return
		}
	}
	storage, err := fs.GetStorage(path, &fs.GetStoragesArgs{})
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if storage.Config().NoUpload {
		common.ErrorStrResp(c, "Current storage doesn't support upload", 405)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	f, err := file.Open()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	defer f.Close()
	dir, name := stdpath.Split(path)
	if shouldIgnoreSystemFile(name) {
		common.ErrorStrResp(c, errs.IgnoredSystemFile.Error(), 403)
		return
	}
	h := make(map[*utils.HashType]string)
	if md5 := c.GetHeader("X-File-Md5"); md5 != "" {
		h[utils.MD5] = md5
	}
	if sha1 := c.GetHeader("X-File-Sha1"); sha1 != "" {
		h[utils.SHA1] = sha1
	}
	if sha256 := c.GetHeader("X-File-Sha256"); sha256 != "" {
		h[utils.SHA256] = sha256
	}
	mimetype := file.Header.Get("Content-Type")
	if len(mimetype) == 0 {
		mimetype = utils.GetMimeType(name)
	}
	s := &stream.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     file.Size,
			Modified: getLastModified(c),
			HashInfo: utils.NewHashInfoByMap(h),
		},
		Reader:       f,
		Mimetype:     mimetype,
		WebPutAsTask: asTask,
	}
	var t task.TaskExtensionInfo
	if asTask {
		s.Reader = struct {
			io.Reader
		}{f}
		t, err = fs.PutAsTask(c.Request.Context(), dir, s)
	} else {
		err = fs.PutDirectly(c.Request.Context(), dir, s)
	}
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if t == nil {
		common.SuccessResp(c)
		return
	}
	common.SuccessResp(c, gin.H{
		"task": getTaskInfo(t),
	})
}

// FsChunkUpload handles uploading a single chunk (replaces upstream version with session verification)
func FsChunkUpload(c *gin.Context) {
	uploadId := c.Query("upload_id")
	indexStr := c.Query("index")
	if uploadId == "" || indexStr == "" {
		common.ErrorStrResp(c, "upload_id and index are required", 400)
		return
	}

	if _, err := strconv.Atoi(indexStr); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	_, chunkDir, err := getAndVerifyChunkSession(c, uploadId)
	if err != nil {
		common.ErrorStrResp(c, err.Error(), 403)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	chunkPath := stdpath.Join(chunkDir, indexStr)
	expectedCRC32 := c.GetHeader("X-Chunk-CRC32")

	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	f, err := os.Open(chunkPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	defer f.Close()

	hasher := crc32.NewIEEE()
	io.Copy(hasher, f)
	actualCRC32 := hex.EncodeToString(hasher.Sum(nil))

	if expectedCRC32 != "" && actualCRC32 != expectedCRC32 {
		f.Close()
		os.Remove(chunkPath)
		common.ErrorStrResp(c, fmt.Sprintf("chunk CRC32 mismatch: client=%s, server=%s", expectedCRC32, actualCRC32), 400)
		return
	}

	common.SuccessResp(c, gin.H{
		"crc32": actualCRC32,
	})
}

// FsChunkMerge streams all chunks into storage (replaces upstream version with zero-copy merge)
func FsChunkMerge(c *gin.Context) {
	var req struct {
		UploadId  string `json:"upload_id"`
		AsTask    bool   `json:"as_task"`
		Overwrite bool   `json:"overwrite"`
		Hash      string `json:"hash"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	sessionData, chunkDir, err := getAndVerifyChunkSession(c, req.UploadId)
	if err != nil {
		common.ErrorStrResp(c, err.Error(), 403)
		return
	}

	reqPath := sessionData["path"].(string)
	path, err := resolveUserPath(c, reqPath)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	if err := checkFileExists(c.Request.Context(), path, req.Overwrite); err != nil {
		common.ErrorStrResp(c, err.Error(), 403)
		return
	}

	totalChunks := int(sessionData["total_chunks"].(float64))

	for i := 0; i < totalChunks; i++ {
		chunkPath := stdpath.Join(chunkDir, strconv.Itoa(i))
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			common.ErrorStrResp(c, "chunk "+strconv.Itoa(i)+" not found", 400)
			return
		}
	}

	dir, name := stdpath.Split(path)
	if err := validateFileName(name); err != nil {
		os.RemoveAll(chunkDir)
		common.ErrorStrResp(c, err.Error(), 403)
		return
	}

	lastModified := time.Now()
	if lm, ok := sessionData["last_modified"].(float64); ok && lm > 0 {
		lastModified = time.UnixMilli(int64(lm))
	}

	s, err := buildMergeStream(chunkDir, name, totalChunks, lastModified, req.Hash)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	s.WebPutAsTask = req.AsTask

	var t task.TaskExtensionInfo
	if req.AsTask {
		t, err = fs.PutAsTask(c.Request.Context(), dir, s)
	} else {
		err = fs.PutDirectly(c.Request.Context(), dir, s)
	}

	if err != nil {
		s.Closers.Close()
		common.ErrorResp(c, err, 500)
		return
	}

	if t == nil {
		common.SuccessResp(c)
		return
	}

	common.SuccessResp(c, gin.H{
		"task": getTaskInfo(t),
	})
}
