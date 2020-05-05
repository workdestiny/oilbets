package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/workdestiny/amlporn/config"
)

const (
	typePNG  = "image/png"
	typeJPEG = "image/jpeg"
	typeJPG  = "image/jpg"
)

// ImageToBase64 is encode from form to base 64
func ImageToBase64(file *multipart.FileHeader) (imageBase64, contentType string, err error) {
	typ := file.Header.Get("Content-Type")
	fp, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer fp.Close()

	img, _, err := image.Decode(fp)
	if err != nil {
		return "", "", err
	}

	buf := new(bytes.Buffer)
	switch typ {
	case typeJPEG, typeJPG:
		err = jpeg.Encode(buf, img, nil)
	case typePNG:
		err = png.Encode(buf, img)
	default:
		err = jpeg.Encode(buf, img, nil)
	}
	if err != nil {
		return "", "", err
	}

	imageBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())

	return imageBase64, typ, nil
}

// GetImageFromFile is encode from form to image file
func GetImageFromFile(fileHeader *multipart.FileHeader) (image.Image, error) {

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	m, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Upload is upload image return (height, width, err)
func Upload(ctx context.Context, bucket *storage.BucketHandle, file *multipart.FileHeader, filename string) (int, int, error) {

	typ := file.Header.Get("Content-Type")
	fp, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer fp.Close()

	w := bucket.Object(filename).NewWriter(ctx)
	w.ContentType = typ
	w.CacheControl = "public, max-age=31536000"
	defer w.Close()

	var img image.Image
	switch typ {
	case typeJPEG, typeJPG:

		img, _, err = image.Decode(fp)
		if err != nil {
			return 0, 0, err
		}

		img = resizeUploadPostImage(img)

		err = jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
		if err != nil {
			return 0, 0, err
		}

		// f, err := file.Open()
		// if err != nil {
		// 	return 0, 0, err
		// }

		// _, err = io.Copy(w, fp)
		// if err != nil {
		// 	return 0, 0, err
		// }

		// img, _, err = image.Decode(f)
		// if err != nil {
		// 	return 0, 0, err
		// }
	default:

		img, _, err = image.Decode(fp)
		if err != nil {
			return 0, 0, err
		}

		img = resizeUploadPostImage(img)

		err = processImage(w, img, typ)
	}
	if err != nil {
		return 0, 0, err
	}

	b := img.Bounds()
	return b.Dy(), b.Dx(), nil
}

//GeneratePostImageName create new Image Name
func GeneratePostImageName(id string) string {
	return "postporn/" + id + "-" + uuid.New().String()
}

func processImage(w io.Writer, img image.Image, typ string) error {

	var err error

	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White),
		image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), img, img.Bounds().Min, draw.Over)

	switch typ {
	case typePNG:
		err = png.Encode(w, dst)
	default:
		err = jpeg.Encode(w, dst, &jpeg.Options{Quality: 100})
	}
	if err != nil {
		return err
	}

	return nil
}

func resizeUploadPostImage(m image.Image) image.Image {
	if m.Bounds().Dx() < config.WidthPostImage {
		return m
	}
	return resize.Resize(config.WidthPostImage, 0, m, resize.Lanczos3)
}
