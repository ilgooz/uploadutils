package uploadutils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ilgooz/ersp"
	"github.com/rakyll/magicmime"

	str "github.com/ilgooz/strings"
)

var imageMime = regexp.MustCompile("^image/")

type Image struct {
	Name, OriginalName string
	Stat               os.FileInfo
}

func Upload(path string, maxMemory int64, w http.ResponseWriter, r *http.Request) (Image, error) {
	image := Image{}
	errRes := ersp.New(w, r)
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return image, err
	}
	defer file.Close()

	randName, err := str.Rand(32)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return image, err
	}

	fileName := fmt.Sprintf("%s_%s", randName, handler.Filename)
	filePath := fmt.Sprintf("%s/%s", path, fileName)

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return image, err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return image, err
	}

	err = magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return image, err
	}

	mimetype, err := magicmime.TypeByFile(filePath)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return image, err
	}

	if ok := imageMime.MatchString(mimetype); !ok {
		errRes.Field("image", "must be valid image")
		errRes.Send(http.StatusBadRequest)
		os.Remove(filePath)
		return image, err
	}

	stat, err := f.Stat()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return image, err
	}

	image.Name = fileName
	image.OriginalName = handler.Filename
	image.Stat = stat

	return image, nil
}

func UploadFile(path string, maxMemory int64, w http.ResponseWriter, r *http.Request) (string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	file, _, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer file.Close()

	randName, err := str.Rand(32)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(path, randName)

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, file)

	return randName, err
}
