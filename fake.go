package ivy

import (
	"bytes"
)

type fakeSource struct {
	buffer *bytes.Buffer
	err    error
	path   string
}

func newFakeSource() *fakeSource {
	return &fakeSource{bytes.NewBuffer(nil), nil, ""}
}

func (s *fakeSource) Load(bucket string, filename string) (*bytes.Buffer, error) {
	return s.buffer, s.err
}

func (s *fakeSource) GetFilePath(bucket string, filename string) string {
	return s.path
}

type fakeCache struct {
	buffer *bytes.Buffer
	err    error
}

func newFakeCache() *fakeCache {
	return &fakeCache{bytes.NewBuffer(nil), nil}
}

func (c *fakeCache) Save(bucket, filename, paramsStr string, file []byte) error {
	return c.err
}

func (c *fakeCache) Load(bucket, filename, paramsStr string) (*bytes.Buffer, error) {
	return c.buffer, c.err
}

func (c *fakeCache) Delete(bucket, filename string) error {
	return c.err
}

func (c *fakeCache) Flush(bucket string) error {
	return c.err
}