package main

import (
	"context"
	"fmt"
	"os"

	"github.com/drummonds/photoprism-go-api/api"
)

func main() {
	fmt.Println(("Demo of connecting to photoprism"))

	host := os.Getenv("PHOTOPRISM_DOMAIN")
	token := os.Getenv("PHOTOPRISM_TOKEN")
	provider := api.NewXAuthProvider(token)

	nc, err := api.NewClientWithResponses(host, api.WithRequestEditorFn(provider.Intercept))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result %+v and err %v\n", nc, err)

	ctx := context.Background()
	// Assuming that there are favourites
	favouriteSearch := "favorite:true"
	searchParams := api.SearchAlbumsParams{Count: 10, Q: &favouriteSearch}
	resp, err := nc.SearchAlbumsWithResponse(ctx, &searchParams)
	if err != nil {
		panic(err)
	}
	fmt.Println("---List of Albums---")
	for _, album := range *resp.JSON200 {
		fmt.Printf("%s %s\n", *album.Title, *album.CreatedAt)
	}

}
