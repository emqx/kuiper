package extensions

import (
	"errors"
	"fmt"
	"github.com/emqx/kuiper/common"
	"github.com/emqx/kuiper/xstream/api"
	"os"
	"path"
	"path/filepath"
	"time"
)

type FileType string

const (
	JSON_TYPE FileType = "json"
)

var fileTypes = map[FileType]bool{
	JSON_TYPE: true,
}

type FileSourceConfig struct {
	FileType FileType `json:"fileType"`
	Path     string   `json:"path"`
	Interval int      `json:"interval"`
}

// The BATCH to load data from file at once
type FileSource struct {
	file   string
	config *FileSourceConfig
}

func (fs *FileSource) Close(ctx api.StreamContext) error {
	ctx.GetLogger().Infof("Close file source")
	// do nothing
	return nil
}

func (fs *FileSource) Configure(fileName string, props map[string]interface{}) error {
	cfg := &FileSourceConfig{}
	err := common.MapToStruct(props, cfg)
	if err != nil {
		return fmt.Errorf("read properties %v fail with error: %v", props, err)
	}
	if cfg.FileType == "" {
		return errors.New("missing or invalid property fileType, must be 'json'")
	}
	if _, ok := fileTypes[cfg.FileType]; !ok {
		return fmt.Errorf("invalid property fileType: %s", cfg.FileType)
	}
	if cfg.Path == "" {
		return errors.New("missing property Path")
	}
	if fileName == "" {
		return errors.New("file name must be specified")
	}
	if !filepath.IsAbs(cfg.Path) {
		cfg.Path, err = common.GetLoc("/" + cfg.Path)
		if err != nil {
			return fmt.Errorf("invalid path %s", cfg.Path)
		}
	}
	fs.file = path.Join(cfg.Path, fileName)

	if fi, err := os.Stat(fs.file); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s not exist", fs.file)
		} else if !fi.Mode().IsRegular() {
			return fmt.Errorf("file %s is not a regular file", fs.file)
		}
	}
	fs.config = cfg
	return nil
}

func (fs *FileSource) Open(ctx api.StreamContext, consumer chan<- api.SourceTuple, errCh chan<- error) {
	err := fs.Load(ctx, consumer)
	if err != nil {
		errCh <- err
		return
	}
	if fs.config.Interval > 0 {
		ticker := time.NewTicker(time.Millisecond * time.Duration(fs.config.Interval))
		logger := ctx.GetLogger()
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				logger.Debugf("Load file source again at %v", common.GetNowInMilli())
				err := fs.Load(ctx, consumer)
				if err != nil {
					errCh <- err
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
}

func (fs *FileSource) Load(ctx api.StreamContext, consumer chan<- api.SourceTuple) error {
	switch fs.config.FileType {
	case JSON_TYPE:
		ctx.GetLogger().Debugf("Start to load from file %s", fs.file)
		resultMap := make([]map[string]interface{}, 0)
		err := common.ReadJsonUnmarshal(fs.file, &resultMap)
		if err != nil {
			return fmt.Errorf("loaded %s, check error %s", fs.file, err)
		}
		ctx.GetLogger().Debug("Sending tuples")
		for _, m := range resultMap {
			select {
			case consumer <- api.NewDefaultSourceTuple(m, nil):
				// do nothing
			case <-ctx.Done():
				return nil
			}
		}
		// Send EOF
		select {
		case consumer <- api.NewDefaultSourceTuple(nil, nil):
			// do nothing
		case <-ctx.Done():
			return nil
		}
		ctx.GetLogger().Debug("All tuples sent")
		return nil
	}
	return fmt.Errorf("invalid file type %s", fs.config.FileType)
}
