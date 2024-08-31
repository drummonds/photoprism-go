package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/drummonds/photoprism-go-api/api"
)

func getClient() *api.ClientWithResponses {
	host := os.Getenv("PHOTOPRISM_DOMAIN")
	token := os.Getenv("PHOTOPRISM_TOKEN")
	provider := api.NewXAuthProvider(token)

	nc, err := api.NewClientWithResponses(host, api.WithRequestEditorFn(provider.Intercept))
	if err != nil {
		panic(err)
	}
	return nc
}

// Search for first album
// then search for first 10 pictures in that album
// then retrun that as a list
func getPhotoList(ctx context.Context) []string {

	searchParams := api.SearchAlbumsParams{Count: 1}
	resp, err := GlobalPage.Client.SearchAlbumsWithResponse(ctx, &searchParams)
	if err != nil {
		panic(err)
	}
	if resp.HTTPResponse.StatusCode != 200 {
		panic(fmt.Errorf("Problem with status %v\n", resp.HTTPResponse.StatusCode))
	}
	if len(*resp.JSON200) < 1 {
		panic("no albums to show")
	}
	album := (*resp.JSON200)[0]
	albumUid := album.UID
	fmt.Printf("%s %s %s\n", *album.Title, *album.CreatedAt, *albumUid)

	// Get photos from album
	photoParams := api.SearchPhotosParams{Count: 10}
	photos, err := GlobalPage.Client.SearchPhotosWithResponse(ctx, &photoParams)
	if err != nil {
		panic(err)
	}
	if photos.HTTPResponse.StatusCode != 200 {
		panic(fmt.Errorf("Problem with status %v\n", photos.HTTPResponse.StatusCode))
	}
	if len(*photos.JSON200) < 1 {
		panic("no photos to show")
	}
	photoList := make([]string, 0, 10)
	for _, photo := range *photos.JSON200 {
		fmt.Printf("%+v\n", photo.OriginalName)
		photoList = append(photoList, *photo.UID)
	}
	return photoList
}

// Downloads global image as temp.jpg
func getImage(ctx context.Context) {
	// Get album Id
	uid := GlobalPhotoList[GlobalPage.PhotoIndex]
	// Get details by search
	// SearchPhotosWithResponse(ctx context.Context, params *SearchPhotosParams, reqEditors ...RequestEditorFn) (*SearchPhotosResponse, error)

	// Get details of photo but hash doesn't seem to work
	photo, err := GlobalPage.Client.GetPhotoWithResponse(ctx, uid)
	if err != nil {
		panic(err)
	}
	files := photo.JSON200.Files
	if len(*files) > 0 {
		fileEntity := (*files)[0]
		switch {
		case true: // Download raw file
			// now get actual data
			hash := *fileEntity.Hash
			// hash := GlobalPhotoList[GlobalPage.PhotoIndex]

			file, err := GlobalPage.Client.GetDownloadWithResponse(ctx, hash)
			if err != nil {
				panic(err)
			}
			status := file.HTTPResponse.StatusCode
			if status != 200 {
				panic(fmt.Errorf("Problem with status downloading file %v\n", file.HTTPResponse.StatusCode))
			}
			os.WriteFile("temp.jpg", file.Body, 0666)
		case true: // Download thumbnail
			hash := *fileEntity.Hash
			token := os.Getenv("PHOTOPRISM_TOKEN")
			file, err := GlobalPage.Client.GetThumbWithResponse(ctx, hash, token, "tile_500")
			if err != nil {
				panic(err)
			}
			status := file.HTTPResponse.StatusCode
			if status != 200 {
				panic(fmt.Errorf("Problem with status downloading file %v\n", file.HTTPResponse.StatusCode))
			}
			os.WriteFile("temp.jpg", file.Body, 0666)
		}
	}
}

type Page struct {
	Title      string
	ImageName  string
	Body       []byte
	Clock      string
	PhotoIndex int
	Client     *api.ClientWithResponses
}

var (
	GlobalPage      = Page{Title: "Album show", ImageName: "TBC"}
	GlobalPhotoList []string
)

func albumHandler(w http.ResponseWriter, r *http.Request) {
	GlobalPage.Clock = time.Now().Format(time.RFC3339)
	err := templates.ExecuteTemplate(w, "album.html", &GlobalPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	getImage(ctx)
	GlobalPage.PhotoIndex = (GlobalPage.PhotoIndex + 1) % len(GlobalPhotoList)
	// http.ServeFile(w, r, "temp.jpg")
	buf, err := os.ReadFile("temp.jpg")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", `attachment;filename="temp.jpg"`)

	w.Write(buf)

}

var templates = template.Must(template.ParseFiles("album.html"))

func main() {
	fmt.Println(("Demo of showing pictures from an album"))
	GlobalPage.Client = getClient()
	ctx := context.Background()
	GlobalPhotoList = getPhotoList(ctx)

	// start web server
	http.HandleFunc("/", albumHandler)
	http.HandleFunc("/image", imageHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
