package streamer

import (
	"bufio"
	"github.com/Mintegral-official/mtggokit/data/container"
	"github.com/easierway/concurrent_map"
	"github.com/pkg/errors"
	"os"
	"runtime"
	"strings"
)

type FileStreamer struct {
	location   int
	container  container.Container
	cfg        *FileStreamerCfg
	f          *os.File
	scan       *bufio.Scanner
	dataParser DataParser
}

type DefaultTextParser struct {
}

func (*DefaultTextParser) Parse(data []byte) (container.MapKey, interface{}, error) {
	s := string(data)
	items := strings.SplitN(s, "\t", 2)
	if len(items) != 2 {
		return nil, errors.New("items len is not 2, item[" + s + "]"), nil
	}
	return concurrent_map.StrKey(items[0]), items[1], nil
}

func DestroyFileStreamer(fs *FileStreamer) {
	_ = fs.f.Close()
}

func NewFileStream(cfg *FileStreamerCfg) (*FileStreamer, error) {
	fs := &FileStreamer{
		cfg:        cfg,
		dataParser: &DefaultTextParser{},
	}
	f, err := os.Open(cfg.Path)
	if err != nil {
		return nil, errors.Wrap(err, "File["+cfg.Path+"]")
	}
	fs.f = f
	runtime.SetFinalizer(fs, DestroyFileStreamer)
	return fs, nil
}

func (fs *FileStreamer) SetProcessor(container container.Container) {
	fs.container = container
}

func (fs *FileStreamer) HasNext() bool {
	return fs.scan.Scan()

}

func (fs *FileStreamer) Next() (container.DataMode, container.MapKey, interface{}, error) {
	k, v, e := fs.dataParser.Parse([]byte(fs.scan.Text()))
	return container.DataModeAdd, k, v, e
}

func (fs *FileStreamer) UpdateData() error {
	switch fs.cfg.Mode {
	case "static":
	case "dynamic":
	case "increase":
		if fs.f != nil {
			_ = fs.f.Close()
		}
		f, err := os.Open(fs.cfg.Path)
		if err != nil {
			return err
		}
		fs.f = f
		_, _ = fs.f.Seek(0, 0)
		return fs.container.LoadBase(fs)
	default:
		return errors.New("not support mode[" + fs.cfg.Mode + "]")
	}

	return nil

}
