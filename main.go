package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	// shell "github.com/stateless-minds/go-ipfs-api"
)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	// sh := shell.NewShell("localhost:5001")

	// err := sh.OrbitDocsDelete(dbCard, "all")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// fmt.Println("completed")

	// folder := "./web/assets/specials/speed"
	// imagesMap, err := scanAndEncodeImages(folder)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	// for filename, image := range imagesMap {
	// 	re := regexp.MustCompile(`^(\d+)\.jpeg$`)
	// 	matches := re.FindStringSubmatch(filename)
	// 	r := rand.New(rand.NewSource(time.Now().UnixNano())) // local random generator with seed
	// 	min := 81
	// 	max := 100
	// 	randomNumber := r.Intn(max-min+1) + min // random number between min and max
	// 	card := Card{
	// 		ID:       matches[1],
	// 		Category: "legs",
	// 		Image:    image,
	// 		Stat:     "speed",
	// 		Value:    randomNumber,
	// 		Special:  "yes",
	// 	}

	// 	cardJSON, err := json.Marshal(card)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		return
	// 	}

	// 	fmt.Printf("Filename: %s\n", filename) // Print a snippet for brevity

	// 	err = sh.OrbitDocsPut(dbCard, cardJSON)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		return
	// 	}

	// 	fmt.Println("completed")
	// }

	// The first thing to do is to associate the hello component with a path.
	//
	// This is done by calling the Route() function,  which tells go-app what
	// component to display for a given path, on both client and server-side.
	app.Route("/", func() app.Composer { return &home{} })
	app.Route("/map", func() app.Composer { return &mapLibre{} })
	app.RouteWithRegexp(`^/delivery/([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$`, func() app.Composer { return &delivery{} })
	app.RouteWithRegexp(`/comp/([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$`, func() app.Composer { return &comp{} })
	app.Route("/report", func() app.Composer { return &report{} })
	app.Route("/not-found", func() app.Composer { return &notFound{} })

	// Once the routes set up, the next thing to do is to either launch the app
	// or the server that serves the app.
	//
	// When executed on the client-side, the RunWhenOnBrowser() function
	// launches the app,  starting a loop that listens for app events and
	// executes client instructions. Since it is a blocking call, the code below
	// it will never be executed.
	//
	// When executed on the server-side, RunWhenOnBrowser() does nothing, which
	// lets room for server implementation without the need for precompiling
	// instructions.
	app.RunWhenOnBrowser()

	// Finally, launching the server that serves the app is done by using the Go
	// standard HTTP package.
	//
	// The Handler is an HTTP handler that serves the client and all its
	// required resources to make it work into a web browser. Here it is
	// configured to handle requests with a path that starts with "/".
	http.Handle("/", &app.Handler{
		Name:        "Cyber Dérive",
		Description: "Deliver for fun.",
		Styles: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.css",
			"https://unpkg.com/maplibre-gl/dist/maplibre-gl.css",
			"web/app.css",
		},
		Scripts: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.js",
			"https://unpkg.com/maplibre-gl/dist/maplibre-gl.js",
			"https://unpkg.com/@maplibre/maplibre-gl-leaflet/leaflet-maplibre-gl.js",
			"web/setupMap.js",
			"web/setupAvatar.js",
			"web/comp.js",
		},
	})

	http.Handle("/map", &app.Handler{
		Name:        "Cyber Dérive",
		Description: "Deliver for fun.",
		Styles: []string{
			"web/app.css",
		},
		Scripts: []string{
			"web/setupMap.js",
		},
	})

	http.Handle("/comp", &app.Handler{
		Name:        "Cyber Dérive",
		Description: "Deliver for fun.",
		Styles: []string{
			"web/app.css",
		},
		Scripts: []string{
			"web/comp.js",
		},
	})

	http.Handle("/report", &app.Handler{
		Name:        "Cyber Dérive",
		Description: "Deliver for fun.",
		Styles: []string{
			"web/app.css",
		},
	})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}

// scanAndEncodeImages scans the given folder for image files and returns a map
// of filename to Base64-encoded image content.
func scanAndEncodeImages(folderPath string) (map[string]string, error) {
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".webp": true,
	}
	result := make(map[string]string)

	// Walk through the directory
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && imageExtensions[strings.ToLower(filepath.Ext(info.Name()))] {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			result[info.Name()] = encoded
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
