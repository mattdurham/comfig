package pkg

import (
	"bytes"
	"fmt"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigMaker(t *testing.T) {
	configStr := `name: bob
`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "test.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)
	fileFS := fmt.Sprintf("file://%s", tDir)
	cmf := NewComfigurator()

	configs, err := cmf.GenerateFromPath(fileFS, "*.yml", func() interface{} {
		return &TestConfig{}
	})

	assert.Nil(t, err)
	assert.True(t, len(configs) == 1)
	found := false
	for k, v := range configs {
		if strings.HasSuffix(k, "test.yml") {
			assert.True(t, v.(*TestConfig).Name == "bob")
			found = true
		}
	}
	assert.True(t, found)
	_ = os.RemoveAll(tDir)
}

func TestConfigMakerWithFakeFiles(t *testing.T) {
	configStr := `name: bob
`

	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "test.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)

	badpath := filepath.Join(tDir, "test.fake")
	err = ioutil.WriteFile(badpath, []byte(configStr), fs.ModePerm)

	assert.Nil(t, err)
	fileFS := fmt.Sprintf("file://%s", tDir)
	cmf := NewComfigurator()

	configs, err := cmf.GenerateFromPath(fileFS, "*.yml", func() interface{} {
		return &TestConfig{}
	})

	assert.Nil(t, err)
	assert.True(t, len(configs) == 1)
	found := false
	for k, v := range configs {
		if strings.HasSuffix(k, "test.yml") {
			assert.True(t, v.(*TestConfig).Name == "bob")
			found = true
		}
	}
	assert.True(t, found)

	_ = os.RemoveAll(tDir)

}

func TestConfigMakerWithMultipleFiles(t *testing.T) {
	configStr := `name: bob
`
	configStr2 := `name: tommy
`
	cmf := NewComfigurator()
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "test.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)

	badpath := filepath.Join(tDir, "test1.yml")
	err = ioutil.WriteFile(badpath, []byte(configStr2), fs.ModePerm)

	assert.Nil(t, err)
	fileFS := fmt.Sprintf("file://%s", tDir)
	configs, err := cmf.GenerateFromPath(fileFS, "*.yml", func() interface{} {
		return &TestConfig{}
	})
	assert.Nil(t, err)
	assert.True(t, len(configs) == 2)

	_ = os.RemoveAll(tDir)
}

func TestS3(t *testing.T) {
	backend := s3mem.New()
	faker := gofakes3.New(backend)

	srv := httptest.NewServer(faker.Server())
	backend.CreateBucket("mybucket")
	t.Cleanup(srv.Close)
	configStr := `name: bob`
	_, err := backend.PutObject(
		"mybucket",
		"test.yml",
		map[string]string{"Content-Type": "application/yaml"},
		bytes.NewBufferString(configStr),
		int64(len(configStr)),
	)

	u, err := url.Parse(srv.URL)
	os.Setenv("AWS_ANON", "true")
	defer os.Unsetenv("AWS_ANON")

	s3Url := "s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host
	assert.NoError(t, err)
	cmf := NewComfigurator()

	configs, err := cmf.GenerateFromPath(s3Url, "*.yml", func() interface{} {
		return &TestConfig{}
	})
	assert.Nil(t, err)
	assert.True(t, len(configs) == 1)
	found := false
	for k, v := range configs {
		if strings.HasSuffix(k, "test.yml") {
			assert.True(t, v.(*TestConfig).Name == "bob")
			found = true
		}
	}
	assert.True(t, found)
}

func TestTemplate(t *testing.T) {
	configStr := `name: {{ .Get "name" }}`
	tDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	fullpath := filepath.Join(tDir, "test.yml")
	err = ioutil.WriteFile(fullpath, []byte(configStr), fs.ModePerm)
	assert.Nil(t, err)
	fileFS := fmt.Sprintf("file://%s", tDir)
	cmf := NewComfigurator()
	kvg := NewKVStoreGateway()
	memstore := NewMemoryStore()
	memstore.Cache["name"] = "bob"
	kvg.AddStore(memstore)
	cmf.AddKVStoreGateway(kvg)

	configs, err := cmf.GenerateFromPath(fileFS, "*.yml", func() interface{} {
		return &TestConfig{}
	})

	assert.Nil(t, err)
	assert.True(t, len(configs) == 1)
	found := false
	for k, v := range configs {
		if strings.HasSuffix(k, "test.yml") {
			assert.True(t, v.(*TestConfig).Name == "bob")
			found = true
		}
	}
	assert.True(t, found)
	_ = os.RemoveAll(tDir)
}

type TestConfig struct {
	Name string `yaml:"name"`
}
