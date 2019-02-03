package main

import (
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"runtime"
	"fmt"
	"strconv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"github.com/jroimartin/gocui"
	"github.com/nfnt/resize"
	"github.com/qeesung/image2ascii/convert"
)
// Your Marvel Credentials goes here
const PRIVATEKEY = "YOUR_MARVEL_PRIVATE_KEY"
const PUBLICKEY = "YOUR_MARVEL_PUBLIC_KEY"
const OPTSEARCH = "1. Search"
const OPTLIST = "2. List"
const OPTEXIT = "Exit [^C]"

type MarvelData struct {
	Code int `json:"code"`
	Status string `json:"status"`
	Copyright string `json:"copyright"`
	AttrTxt string `json:"attributionText"`
	AttrHtml string `json:"attributionHTML"`
	Etag string `json:"etag"`
	Datos Data `json:"data"`
}
type Data struct {
	Offset int `json:"offset"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Count int `json:"count"`
	Results []Result `json:"results"`
}

type Result struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Modified int64 `json:"modified"`
	Picture Thumbnail `json:"thumbnail"`
	ResUri string `json:"resourceURI"`
	Comics Comic `json:"comics"`
}

type Thumbnail struct {
	Path string `json:"path"`
	Ext string `json:"extension"`
}

type Comic struct {
	Available int `json:"available"`
	CollectionUri string `json:"collectionURI"`
	Items []ComicItem `json:"items"`
	Returned int `json:"returned"`
}

type ComicItem struct {
	ResUri string `json:"resourceURI"`
	Name string `json:"name"`
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}


}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "side" {
		_ , err := g.SetCurrentView("main")
		return err
	}
	_, err := g.SetCurrentView("side")
	return err
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		var l string
		var err error
		cx, cy := v.Cursor()
		if l, err = v.Line(cy+1); err != nil {
			l = ""
		}
		if (l == "") {
			return nil
		} else{

			if err := v.SetCursor(cx, cy+1); err != nil {
				ox, oy := v.Origin()
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy -1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy -1 ); err != nil {
				return err
			}
		}
	}
	return nil
}

func DownloadFile(filePath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the bytes to a file
	_, err = io.Copy(out, resp.Body)
	return err
}

func search(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error
	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	if l != "" {

		ts := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
		hash := md5.Sum([]byte(ts+PRIVATEKEY+PUBLICKEY))
		hashed := hex.EncodeToString(hash[:])
		var Url *url.URL
		Url, err := url.Parse("http://gateway.marvel.com/")
		Url.Path += "/v1/public/characters"
		parameters := url.Values{}
		parameters.Add("ts", ts)
		parameters.Add("apikey",PUBLICKEY)
		parameters.Add("hash",hashed)
		if l == OPTLIST {
			parameters.Add("orderBy", "name")
			parameters.Add("limit", "20")
		}else {
			parameters.Add("name", l)
		}
		Url.RawQuery = parameters.Encode()
		res, err := http.Get(Url.String())
		if err != nil {
			log.Fatal(err)
		}
		robots, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		var result MarvelData
		json.Unmarshal([]byte(robots), &result)

		g.Update(func(g *gocui.Gui) error {
			v, err := g.View("main")
			if err != nil {
				return err
			}
			v.Clear()
			if l == OPTLIST {
				fmt.Fprintf(v, "List of first 20 Marvel characters:\n\n")
				for i := 0; i < len(result.Datos.Results); i++ {
					fmt.Fprintf(v, "%d.- %s\n",i+1, result.Datos.Results[i].Name)
				}
			} else {

				fmt.Fprintf(v, "Character: %s\n", result.Datos.Results[0].Name)
				fmt.Fprintf(v, "\nDescription:\n%s\n", result.Datos.Results[0].Description)
				fmt.Fprintf(v, "\nURI: %s\n", result.Datos.Results[0].ResUri)
				picUrl := fmt.Sprintf("%s.%s", result.Datos.Results[0].Picture.Path, result.Datos.Results[0].Picture.Ext)
				fmt.Fprintf(v, "\nPicture: %s\n", picUrl)
				fmt.Fprintf(v, "\nComics: %d\n", result.Datos.Results[0].Comics.Available)
				if err := DownloadFile("thumb.jpg", picUrl); err != nil {
					return err
				}
				file, err := os.Open("thumb.jpg")
				if err != nil {
					return err
				}
				img, err := jpeg.Decode(file)
				if err != nil {
					log.Fatal(err)
				}
				file.Close()

				m := resize.Resize(128, 0, img, resize.Lanczos3)
				out, err := os.Create("thumbsm.jpg")
				if err != nil {
					return err
				}
				defer out.Close()
				jpeg.Encode(out, m, nil)

				convertOptions := convert.DefaultOptions
				convertOptions.FixedWidth = 95
				convertOptions.FixedHeight = 70
				// convertOptions.FitScreen = true
				convertOptions.Colored = true
				converter := convert.NewImageConverter()
				fmt.Fprintf(v, "\n%s",converter.ImageFile2ASCIIString("thumbsm.jpg", &convertOptions))
			}

			return nil
		})
	}

	if err := g.DeleteView("msg"); err != nil {
		return err
	}

	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}
	return nil
}

func getLine(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error
	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	i := strings.Index(l, "1.")
	if i > -1 {
		maxX, maxY := g.Size()
		if v, err := g.SetView("msg", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Escriba el super heroe a buscar"
			fmt.Fprintln(v, "")
			v.Editable = true
			if _, err := g.SetCurrentView("msg"); err != nil {
				return err
			}
		}
		return nil
	}
	i = strings.Index(l, "2.")
	if i>-1 {
		search(g, v)

	}
	i = strings.Index(l, "Exit")
	if i>-1 {
		if runtime.GOOS == "windows" {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()
		} else {
			fmt.Println("\033[2J")
		}
		os.Exit(0)
	}

	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("side", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyEnter, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		return err
	}
	if err := g.SetKeybinding("msg", gocui.KeyEnter, gocui.ModNone, search); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	return nil
}

func layout (g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("side", 1, 1, 15, maxY-1); err != nil{
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Title = "Marvel [Tab]"
		fmt.Fprintln(v, OPTSEARCH)
		fmt.Fprintln(v, OPTLIST)
		fmt.Fprintln(v, OPTEXIT)
	}
	if v, err := g.SetView("main", 	16, 1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "                               | |")
		fmt.Fprintln(v, " _ __ ___   __ _ _ ____   _____| |")
	    fmt.Fprintln(v, "| '_ ` _ \\ / _` | '__\\ \\ / / _ \\ |")
	    fmt.Fprintln(v, "| | | | | | (_| | |   \\ V /  __/ |")
		fmt.Fprintln(v, "|_| |_| |_|\\__,_|_|    \\_/ \\___|_|")
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "Marvel API demo by @vhanla")
		v.Wrap = true
		v.Editable = true

		if _, err := g.SetCurrentView("side"); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}