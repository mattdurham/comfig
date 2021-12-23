package pkg

import (
	"bytes"
	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/blobfs"
	"github.com/hairyhenderson/go-fsimpl/filefs"
	"github.com/hairyhenderson/go-fsimpl/gitfs"
	"github.com/hairyhenderson/go-fsimpl/httpfs"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"path/filepath"
	"text/template"
)

type Comfigurator struct {
	mux fsimpl.FSMux
	kvg *KVStoreGateway
}

func NewComfigurator() *Comfigurator {
	cmfg := &Comfigurator{}
	cmfg.mux = newFSProvider()
	cmfg.kvg = NewKVStoreGateway()
	return cmfg
}

func newFSProvider() fsimpl.FSMux {
	mux := fsimpl.NewMux()
	mux.Add(filefs.FS)
	mux.Add(httpfs.FS)
	mux.Add(blobfs.FS)
	mux.Add(gitfs.FS)
	return mux
}

func (c *Comfigurator) AddKVStoreGateway(kvg *KVStoreGateway) {
	c.kvg = kvg
}

// GenerateFromPath creates a series of yaml configs based on the path and a given string pattern.
// the pattern is the same as used by filepath.Match.
func (c *Comfigurator) GenerateFromPath(path string, pattern string, configMake func() interface{}) (map[string]interface{}, error) {
	handler, err := c.mux.Lookup(path)
	if err != nil {
		return nil, err
	}
	files, err := fs.ReadDir(handler, ".")
	if err != nil {
		return nil, err
	}
	configs := make(map[string]interface{}, 0)
	for _, f := range files {
		// We don't walk directories
		if f.IsDir() {
			continue
		}
		matched, _ := filepath.Match(pattern, f.Name())
		if matched {
			file, err := handler.Open(f.Name())
			stats, err := f.Info()
			if err != nil {
				return nil, err
			}
			fBytes := make([]byte, stats.Size())
			bytesRead, err := file.Read(fBytes)
			if bytesRead == 0 {
				continue
			}

			if err != nil && err != io.EOF {
				return nil, err
			}
			fString := string(fBytes)
			// Parse the template
			tmp, err := template.New("t").Parse(fString)
			if err != nil {
				return nil, err
			}
			buf := new(bytes.Buffer)
			err = tmp.Execute(buf, c.kvg)
			if err != nil {
				return nil, err
			}
			cfg := configMake()
			err = yaml.Unmarshal(buf.Bytes(), cfg)
			if err != nil && err != io.EOF {
				return nil, err
			}
			fullname := filepath.Join(path, f.Name())
			configs[fullname] = cfg
		}
	}
	return configs, nil
}
