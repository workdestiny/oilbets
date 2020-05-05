package app

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"regexp"

	"github.com/google/uuid"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	"github.com/workdestiny/amlporn/config"
)

func resizeImage(m image.Image) (image.Image, int, int) {

	b := m.Bounds()
	return m, b.Dy(), b.Dx()

}

func resizeTopicImage(m image.Image) image.Image {
	return imaging.Resize(m, config.WidthTopic, config.HeightTopic, imaging.Lanczos)
}

func resizeUploadDisplayImage(m image.Image) (image.Image, image.Image, image.Image) {

	return imaging.Resize(m, config.WidthDisplayMini, config.HeightDisplayMini, imaging.Lanczos),
		imaging.Resize(m, config.WidthDisplayMiddle, config.HeightDisplayMiddle, imaging.Lanczos),
		imaging.Resize(m, config.WidthDisplay, config.HeightDisplay, imaging.Lanczos)
}

func resizeUploadCoverImage(m image.Image) (image.Image, image.Image) {

	ms := resize.Resize(0, config.HeightResizeCoverMini, m, resize.Lanczos3)
	b := ms.Bounds()
	width := b.Dx()
	if width > config.WidthResizeCoverMini {
		width = config.WidthResizeCoverMini
	}

	return imaging.CropCenter(m, width, config.HeightResizeCoverMini), imaging.Resize(m, config.WidthCover, config.HeightCover, imaging.Lanczos)
}

func generateGapDisplayName(id string) (string, string, string) {
	return "gap/display-mini/" + id + "-" + uuid.New().String(), "gap/display-middle/" + id + "-" + uuid.New().String(), "gap/display/" + id + "-" + uuid.New().String()
}

func generatProfileDisplayName(id string) (string, string, string) {
	return "profile/display-mini/" + id + "-" + uuid.New().String(), "profile/display-middle/" + id + "-" + uuid.New().String(), "profile/display/" + id + "-" + uuid.New().String()
}

func generateGapCoverName(id string) (string, string) {
	return "gap/cover-mini/" + id + "-" + uuid.New().String(), "gap/cover/" + id + "-" + uuid.New().String()
}

func generateTopicName(id string) string {
	return "topic/normal/" + id + "-" + uuid.New().String()
}

func generateProfileBookbankName(id string) string {
	return "profile/bookbank/" + id + "-" + uuid.New().String()
}

func resizeDisplayImage(m image.Image) image.Image {
	b := m.Bounds()
	width := b.Dx()
	height := b.Dy()
	if width <= config.HeightDisplay && height <= config.WidthDisplay {
		return m
	}

	return imaging.Fill(m, config.HeightDisplay, config.WidthDisplay, imaging.Center, imaging.Lanczos)
	// return imaging.Resize(m, 300, 0, imaging.Lanczos)
}

func resizeMainImage(m image.Image) (image.Image, int, int) {

	editedImage := m

	b := m.Bounds()
	width := b.Dx()
	height := b.Dy()

	if width > config.WidthMainPost {

		editedImage = resize.Resize(config.WidthMainPost, 0, m, resize.Lanczos3)
		be := editedImage.Bounds()
		if be.Dy() > config.HigthMainPost {
			return imaging.Fill(m, config.WidthMainPost, config.HigthMainPost, imaging.Top, imaging.Lanczos),
				config.WidthMainPost, config.HigthMainPost
		}
		return editedImage, config.WidthMainPost, be.Dy()
	}

	if height > config.HigthMainPost {
		return imaging.Fill(m, width, config.HigthMainPost, imaging.Top, imaging.Lanczos),
			b.Dx(), config.HigthMainPost
	}

	return m, b.Dx(), b.Dy()

}

func resizeMainImageMobile(m image.Image) (image.Image, int, int) {

	editedImage := m

	b := m.Bounds()
	width := b.Dx()
	height := b.Dy()

	if width > config.WidthMainPostMobile {

		editedImage = resize.Resize(config.WidthMainPostMobile, 0, m, resize.Lanczos3)
		be := editedImage.Bounds()
		if be.Dy() > config.HigthMainPostMobile {
			return imaging.Fill(m, config.WidthMainPostMobile, config.HigthMainPostMobile, imaging.Top, imaging.Lanczos),
				config.WidthMainPostMobile, config.HigthMainPostMobile
		}
		return editedImage, config.WidthMainPostMobile, be.Dy()
	}

	if height > config.HigthMainPostMobile {
		return imaging.Fill(m, width, config.HigthMainPostMobile, imaging.Top, imaging.Lanczos),
			b.Dx(), config.HigthMainPostMobile
	}

	return m, b.Dx(), b.Dy()

}

func resizefacebookImage(m image.Image) (image.Image, int, int) {

	mfb := resize.Resize(0, config.HigthImageFacebook, m, resize.Lanczos3)
	return imaging.Fill(mfb, config.WidthImageFacebook, config.HigthImageFacebook, imaging.Top, imaging.Lanczos),
		config.WidthImageFacebook, config.HigthImageFacebook
	// be := mfb.Bounds()
	// if be.Dx() > config.WidthImageFacebook {
	// 	return imaging.Fill(mfb, config.WidthImageFacebook, config.HigthImageFacebook, imaging.Top, imaging.Lanczos),
	// 		config.WidthImageFacebook, config.HigthImageFacebook
	// }
	// return mfb, be.Dx(), be.Dy()
}

func resizeDisplayMiniImage(m image.Image) image.Image {

	return imaging.Resize(m, config.HeightDisplayMini, config.WidthDisplayMini, imaging.Lanczos)
	// return imaging.Resize(m, 300, 0, imaging.Lanczos)
}

func generateDownloadURL(filename string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket.Name, filename)
}

func generateDisplayImageName(id string) string {
	return "profile/display/" + id + "-" + uuid.New().String()
}

func generateDisplayMiniImageName(id string) string {
	return "profile/display-mini/" + id + "-" + uuid.New().String()
}

func generateMainImagePostName(id string) string {
	return "postp/main/" + id + "-" + uuid.New().String()
}

func generateMainImagePostNameMobile(id string) string {
	return "postp/main/mobile/" + id + "-" + uuid.New().String()
}

func encodeJPEG(w io.Writer, m image.Image, quality int) error {
	return jpeg.Encode(w, m, &jpeg.Options{Quality: quality})
}

func encodePNG(w io.Writer, m image.Image) error {
	return png.Encode(w, m)
}

func upload(ctx context.Context, m image.Image, filename string) error {

	writer := bucket.Storage.Object(filename).NewWriter(ctx)
	writer.CacheControl = "public, max-age=31536000"
	defer writer.Close()
	err := encodeJPEG(writer, m, 80)
	if err != nil {
		return err
	}
	return nil
}

func uploadThumbnailMobile(ctx context.Context, m image.Image, filename string) error {

	writer := bucket.Storage.Object(filename).NewWriter(ctx)
	writer.CacheControl = "public, max-age=31536000"
	defer writer.Close()
	err := encodeJPEG(writer, m, 80)
	if err != nil {
		return err
	}
	return nil
}

var imgRE = regexp.MustCompile(`<img[^>]+\bsrc=["']([^"']+)["']`)

// if your img's are properly formed with doublequotes then use this, it's more efficient.
// var imgRE = regexp.MustCompile(`<img[^>]+\bsrc="([^"]+)"`)
func findImages(htm string) []string {
	imgs := imgRE.FindAllStringSubmatch(htm, -1)
	out := make([]string, len(imgs))
	for i := range out {
		out[i] = imgs[i][1]
	}
	return out
}
