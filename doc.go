package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/coalaura/lock"
	"github.com/gen2brain/go-fitz"
	"github.com/gen2brain/webp"
	"github.com/go-chi/chi/v5"
)

var (
	docRgx = regexp.MustCompile(`^[a-zA-Z0-9_-]{25,50}$`)
	locks  = lock.NewLockMap[string]()
)

func handleDoc(w http.ResponseWriter, r *http.Request) {
	doc := chi.URLParam(r, "doc")
	if !docRgx.MatchString(doc) {
		w.WriteHeader(http.StatusBadRequest)

		log.Warnln("doc: invalid document id")

		return
	}

	var page int

	if raw := chi.URLParam(r, "page"); raw != "" {
		pg, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			log.Warnln("doc: invalid page number")
			log.Warnln(err)

			return
		}

		page = int(pg)
	}

	page = min(10, max(1, page))

	img, code, err := downloadDocAsPNG(doc, page)
	if err != nil {
		if code == 0 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(code)
		}

		log.Warnln("doc: failed to download doc")
		log.Warnln(err)

		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusOK)

	w.Write(img)
}

func downloadDocAsPNG(doc string, page int) ([]byte, int, error) {
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		os.MkdirAll("docs", 0755)
	}

	locks.Lock(doc)
	defer locks.Unlock(doc)

	path := filepath.Join("docs", fmt.Sprintf("%s.%d.webp", doc, page))

	var buf bytes.Buffer

	if _, err := os.Stat(path); os.IsNotExist(err) {
		data, code, err := downloadDocAsPdf(doc)
		if err != nil {
			return nil, code, err
		}

		pdf, err := fitz.NewFromMemory(data)
		if err != nil {
			return nil, 0, err
		}

		if pdf.NumPage() < page {
			return nil, http.StatusBadRequest, fmt.Errorf("page %d does not exist", page)
		}

		img, err := pdf.ImageDPI(page-1, 300)
		if err != nil {
			return nil, 0, err
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return nil, 0, err
		}

		defer file.Close()

		wr := io.MultiWriter(&buf, file)

		err = webp.Encode(wr, img, webp.Options{
			Quality: 90,
		})

		if err != nil {
			return nil, 0, err
		}
	} else {
		now := time.Now()

		os.Chtimes(path, now, now)

		file, err := os.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return nil, 0, err
		}

		defer file.Close()

		_, err = io.Copy(&buf, file)
		if err != nil {
			return nil, 0, err
		}
	}

	return buf.Bytes(), 0, nil
}

func downloadDocAsPdf(doc string) ([]byte, int, error) {
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		os.MkdirAll("docs", 0755)
	}

	path := filepath.Join("docs", doc+".pdf")

	var buf bytes.Buffer

	if _, err := os.Stat(path); os.IsNotExist(err) {
		uri := fmt.Sprintf("https://docs.google.com/document/d/%s/export?format=pdf", doc)

		resp, err := http.Get(uri)
		if err != nil {
			return nil, 0, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, resp.StatusCode, fmt.Errorf("status %d", resp.StatusCode)
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return nil, 0, err
		}

		defer file.Close()

		wr := io.MultiWriter(&buf, file)

		_, err = io.Copy(wr, resp.Body)
		if err != nil {
			return nil, 0, err
		}
	} else {
		now := time.Now()

		os.Chtimes(path, now, now)

		file, err := os.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return nil, 0, err
		}

		defer file.Close()

		_, err = io.Copy(&buf, file)
		if err != nil {
			return nil, 0, err
		}
	}

	return buf.Bytes(), 0, nil
}
