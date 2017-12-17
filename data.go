package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/termie/go-shutil"
)

type Blawg struct {
	Path   string
	Config *Config
	Posts  []PostLog
}

type PostLog struct {
	Title                string
	Author               string
	Path                 string
	LastUpdated, Created time.Time
	Body                 string
}

func LoadBlawgData(path string) (*Blawg, error) {
	f, err := os.Open(path + "/blawg.json")
	if err != nil {
		return nil, err
	}

	bl := new(Blawg)
	bl.Path = path
	bl.Config = new(Config)

	err = json.NewDecoder(f).Decode(bl.Config)
	if err != nil {
		return nil, err
	}

	for _, v := range bl.Config.Posts {
		fe, err := os.Open(path + "/posts/" + v.Path)
		if err != nil {
			return nil, err
		}

		st, _ := fe.Stat()
		b := new(bytes.Buffer)
		io.Copy(b, fe)

		pl := PostLog{
			Title:       v.Title,
			Path:        v.Path,
			Body:        b.String(),
			Created:     time.Unix(0, v.Timestamp*1000000),
			LastUpdated: st.ModTime(),
			Author:      v.Author,
		}

		bl.Posts = append(bl.Posts, pl)
	}

	return bl, nil
}

type HPost struct {
	Header                         string
	Link                           string
	Author                         string
	Created, LastUpdated           string
	CreatedStamp, LastUpdatedStamp int64
}

type HPostList struct {
	BlogTitle string
	PostList  []HPost
}

func (h *HPostList) Len() int {
	return len(h.PostList)
}

func (h *HPostList) Swap(i, j int) {
	eli := h.PostList[i]
	elj := h.PostList[j]
	h.PostList[i] = elj
	h.PostList[j] = eli
}

func (h *HPostList) Less(i, j int) bool {
	return h.PostList[i].CreatedStamp > h.PostList[j].CreatedStamp
}

type HHeader struct {
	BaseURL string
	Title   string
}

func (b *Blawg) BuildHeader(s string) string {
	t, err := template.New("list").Parse(s)
	if err != nil {
		panic(err)
	}

	o := new(bytes.Buffer)
	t.Execute(o, &HHeader{
		Title:   b.Config.Title,
		BaseURL: b.Config.BaseURL,
	})

	return o.String()
}

func (b *Blawg) BuildList(s string) string {
	t, err := template.New("list").Parse(s)
	if err != nil {
		panic(err)
	}

	h := new(HPostList)
	for _, v := range b.Posts {
		h.PostList = append(h.PostList, b.GetHPost(v))
	}

	sort.Sort(h)

	h.BlogTitle = b.Config.Title

	buf := new(bytes.Buffer)
	t.Execute(buf, h)
	return buf.String()
}

func (b *Blawg) GetHPost(v PostLog) HPost {
	return HPost{
		Header:           v.Title,
		Author:           v.Author,
		LastUpdatedStamp: v.LastUpdated.UnixNano() / 1000000,
		CreatedStamp:     v.Created.UnixNano() / 1000000,
		LastUpdated:      printTime(v.LastUpdated),
		Created:          printTime(v.Created),
		Link:             b.Config.BaseURL + "/posts/" + v.Path,
	}
}

func printTime(t time.Time) string {
	return fmt.Sprintf("%s, %d %s %d", t.Weekday(), t.Day(), t.Month(), t.Year())
}

func (b *Blawg) BuildPostHeader(s string, lg PostLog) string {
	t, err := template.New("phead").Parse(s)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	t.Execute(buf, b.GetHPost(lg))
	return buf.String()
}

func getDir() string {
	o := "/src/github.com/superp00t/blawg/"
	if e := os.Getenv("GOPATH"); e != "" {
		return e + o
	}

	return os.Getenv("HOME") + "/go" + o
}

func (b *Blawg) BuildWeb(out string) error {
	RemoveContents(out)
	shutil.CopyTree(b.Config.Theme+"/assets/", out, nil)
	os.Mkdir(out+"/posts/", 0777)

	thm := getDir() + b.Config.Theme
	headb, err := ioutil.ReadFile(thm + "/header.html")
	if err != nil {
		return err
	}

	pheadb, err := ioutil.ReadFile(thm + "/postheader.html")
	if err != nil {
		return err
	}

	listb, err := ioutil.ReadFile(thm + "/list.html")
	if err != nil {
		return err
	}

	footb, err := ioutil.ReadFile(thm + "/footer.html")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(out+"/index.html", []byte(
		b.BuildHeader(string(headb))+
			b.BuildList(string(listb))+
			string(footb)), 0777)

	if err != nil {
		return err
	}

	for _, v := range b.Posts {
		err = ioutil.WriteFile(out+"/posts/"+v.Path, []byte(
			b.BuildHeader(string(headb))+
				b.BuildPostHeader(string(pheadb), v)+
				v.Body+
				string(footb)), 0777)
		if err != nil {
			return err
		}

	}

	return nil
}
