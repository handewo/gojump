package proxy

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/log"
)

const (
	dateTimeFormat = "20060102"

	replayFilenameSuffix   = ".cast"
	replayGzFilenameSuffix = ".gz"
)

func NewReplayRecord(sid, user, asset string, info *ReplyInfo) (*ReplyRecorder, error) {
	recorder := &ReplyRecorder{
		SessionID:      sid,
		info:           info,
		disableRecorer: config.GetConf().DisableRecorder,
	}

	if recorder.isDisableRecorder() {
		log.Info.Print("replay recorder is disabled")
		return recorder, nil
	}
	today := info.TimeStamp.Format(dateTimeFormat)
	replayRootDir := config.GetConf().ReplayFolderPath
	sessionReplayDirPath := filepath.Join(replayRootDir, today)
	err := common.EnsureDirExist(sessionReplayDirPath)
	if err != nil {
		log.Error.Printf("Create dir %s error: %s", sessionReplayDirPath, err)
		recorder.err = err
		return recorder, err
	}
	filename := user + "_" + asset + "_" + info.TimeStamp.Format("150405") + "_" + sid[:8] + replayFilenameSuffix
	gzFilename := filename + replayGzFilenameSuffix
	absFilePath := filepath.Join(sessionReplayDirPath, filename)
	absGZFilePath := filepath.Join(sessionReplayDirPath, gzFilename)
	storageTargetName := strings.Join([]string{today, gzFilename}, "/")
	recorder.absGzipFilePath = absGZFilePath
	recorder.absFilePath = absFilePath
	recorder.Target = storageTargetName
	fd, err := os.Create(recorder.absFilePath)
	if err != nil {
		log.Error.Printf("Create replay file %s error: %s", recorder.absFilePath, err)
		recorder.err = err
		return recorder, err
	}
	log.Info.Printf("Create replay file %s", recorder.absFilePath)
	recorder.file = fd

	options := make([]common.AsciiOption, 0, 3)
	options = append(options, common.WithHeight(info.Height))
	options = append(options, common.WithWidth(info.Width))
	options = append(options, common.WithTimestamp(info.TimeStamp))
	recorder.Writer = common.NewWriter(recorder.file, options...)
	return recorder, nil
}

type ReplyRecorder struct {
	SessionID string
	info      *ReplyInfo

	absFilePath     string
	absGzipFilePath string
	Target          string
	Writer          *common.AsciiWriter
	err             error

	file *os.File
	once sync.Once

	disableRecorer bool
}

func (r *ReplyRecorder) isDisableRecorder() bool {
	return r.disableRecorer
}

func (r *ReplyRecorder) Record(p []byte) {
	if r.isDisableRecorder() {
		return
	}
	if len(p) > 0 {
		r.once.Do(func() {
			if err := r.Writer.WriteHeader(); err != nil {
				log.Error.Printf("Session %s write replay header failed: %s", r.SessionID, err)
			}
		})
		if err := r.Writer.WriteRow(p); err != nil {
			log.Error.Printf("Session %s write replay row failed: %s", r.SessionID, err)
		}
	}
}

func (r *ReplyRecorder) End() {
	if r.isDisableRecorder() {
		return
	}
	_ = r.file.Close()
	go r.compressReplay()
}

func (r *ReplyRecorder) compressReplay() {
	if !common.FileExists(r.absFilePath) {
		log.Info.Print("Replay file not found, passed: ", r.absFilePath)
		return
	}
	if stat, err := os.Stat(r.absFilePath); err == nil && stat.Size() == 0 {
		log.Info.Print("Replay file is empty, removed: ", r.absFilePath)
		_ = os.Remove(r.absFilePath)
		return
	}
	if !common.FileExists(r.absGzipFilePath) {
		log.Debug.Print("Compress replay file: ", r.absFilePath)
		_ = common.GzipCompressFile(r.absFilePath, r.absGzipFilePath)
		_ = os.Remove(r.absFilePath)
	}
}

type ReplyInfo struct {
	Width     int
	Height    int
	TimeStamp time.Time
}
