package ivy

import (
	"bytes"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

var ErrNotFound = errors.New("not found")

type Source interface {
	Load(bucket string, filename string) (*bytes.Buffer, error)
	GetFilePath(bucket string, filename string) string
}

type Cache interface {
	Save(bucket, filename, paramsStr string, file []byte) error
	Load(bucket, filename, paramsStr string) (*bytes.Buffer, error)
	Delete(bucket, filename string) error
	Flush() error
}

type Ivy struct {
	Source Source
	Cache  Cache
	Config *Config
}

type Config struct {
	HttpCache     int64
	IsDevelopment bool
}

func New(source Source, cache Cache, config *Config) *Ivy {
	return &Ivy{source, cache, config}
}

func (iv *Ivy) Get(bucket, paramsStr, path string, w http.ResponseWriter, r *http.Request) {
	if iv.isNotModify(r) {
		iv.writeNotModify(w, path)
		return
	}

	params, err := parseParams(paramsStr)
	if err != nil {
		iv.writeError(w, err)
		return
	}

	var img *bytes.Buffer

	if params.IsDefault {
		if img, err = iv.loadFromSource(bucket, path, params); err != nil {
			iv.writeError(w, err)
			return
		}
	} else {
		if img, err = iv.loadFromCache(bucket, path, params); err != nil {
			if img, err = iv.loadFromSource(bucket, path, params); err != nil {
				iv.writeError(w, err)
				return
			}
		}
	}

	iv.writeSuccess(w, path, img)
}

func (f *Ivy) isNotModify(r *http.Request) bool {
	if f.Config.HttpCache > 0 && !f.Config.IsDevelopment && r.Header.Get("If-Modified-Since") != "" {
		return true
	}

	return false
}

func (f *Ivy) loadFromCache(bucket, filePath string, params *Params) (*bytes.Buffer, error) {
	if f.Config.IsDevelopment || f.Cache == nil {
		return nil, errors.New("no cache")
	}

	file, err := f.Cache.Load(bucket, filePath, params.String())
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (f *Ivy) loadFromSource(bucket, filePath string, params *Params) (*bytes.Buffer, error) {
	file, err := f.Source.Load(bucket, filePath)
	if err != nil {
		return nil, err
	}

	img, err := process(params, f.Source.GetFilePath(bucket, filePath), file)
	if err != nil {
		return nil, err
	}

	if f.Cache != nil {
		go f.Cache.Save(bucket, filePath, params.String(), img.Bytes())
	}

	return img, nil
}

func (f *Ivy) FlushCache() error {
	return f.Cache.Flush()
}

func (f *Ivy) getContentType(filePath string) string {
	switch filepath.Ext(filePath) {
	case ".jpg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	}

	return "application/octet-stream"
}

func (f *Ivy) setHeader(w http.ResponseWriter, filePath string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", f.getContentType(filePath))
	w.Header().Add("Connection", "keep-alive")
	w.Header().Add("Vary", "Accept-Encoding")
	w.Header().Add("Last-Modified", "Tue, 01 Jan 2008 00:00:00 GMT")

	if f.Config.HttpCache > 0 && !f.Config.IsDevelopment {
		w.Header().Add("Cache-Control", "public; max-age="+strconv.FormatInt(f.Config.HttpCache, 10))
		w.Header().Add("Expires", time.Now().Add(time.Second*time.Duration(f.Config.HttpCache)).Format("Mon, _2 Jan 2006 15:04:05 MST"))
	}
}

func (f *Ivy) writeSuccess(w http.ResponseWriter, filePath string, file *bytes.Buffer) {
	f.setHeader(w, filePath)

	w.WriteHeader(http.StatusOK)
	w.Write(file.Bytes())
}

func (f *Ivy) writeError(w http.ResponseWriter, err error) {
	switch err {
	case ErrNotFound:
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	default:
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}

func (f *Ivy) writeNotModify(w http.ResponseWriter, filePath string) {
	f.setHeader(w, filePath)
	w.WriteHeader(304)
	w.Write(nil)
}
