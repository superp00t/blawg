package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/ogier/pflag"
)

type Config struct {
	BaseURL string `json:"baseURL"`
	Title   string `json:"title"`
	Theme   string `json:"theme"`
	Posts   []Post `json:"posts"`
}

type Post struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
}

var (
	dir = pflag.StringP("path", "p", "path", "path to blog data")
	out = pflag.StringP("out", "o", "", "output dir")
)

func fatal(e string) {
	fmt.Println(e)
	os.Exit(-1)
}

func main() {
	pflag.Parse()

	if pflag.Arg(0) == "newblog" {
		str := `{
		"title": "Regular old blog",
		"theme": "theme/default",
		"baseURL": "/",
		"posts": []
	}`
		err := os.Mkdir(*dir, 0777)
		if err != nil {
			fatal(err.Error())
		}

		err = ioutil.WriteFile(*dir+"/blawg.json", []byte(str), 0777)
		if err != nil {
			fatal(err.Error())
		}

		err = os.Mkdir(*dir+"/posts", 0777)
		if err != nil {
			fatal(err.Error())
		}

		fmt.Println("blog", *dir, "created")
		return
	}

	if pflag.Arg(0) == "newpost" {
		if len(pflag.Args()) < 3 {
			fatal("usage: newpost <author> <title>")
		}

		meta := *dir + "/blawg.json"
		f, err := os.Open(meta)
		if err != nil {
			fatal(err.Error())
		}
		var c Config
		err = json.NewDecoder(f).Decode(&c)
		if err != nil {
			fatal(err.Error())
		}
		h := htmlPath(pflag.Arg(2))
		c.Posts = append(c.Posts, Post{
			Author:    pflag.Arg(1),
			Title:     pflag.Arg(2),
			Path:      h,
			Timestamp: time.Now().UnixNano() / 1000000,
		})
		dat, _ := json.MarshalIndent(c, "", "  ")
		ioutil.WriteFile(meta, dat, 0777)
		pth := *dir + "/posts/" + h
		os.Create(pth)
		fmt.Println(pth, "created")
		return
	}

	b, err := LoadBlawgData(*dir)
	if err != nil {
		fatal(err.Error())
	}

	if *out == "" {
		fatal("Must have output directory -o")
	}

	err = b.BuildWeb(*out)
	if err != nil {
		fatal(err.Error())
	}
}

func htmlPath(input string) string {
	input = strings.Replace(input, " ", "_", -1)
	input = strings.ToLower(input)
	return input + ".html"
}
