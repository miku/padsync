package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/miku/clam"
	"github.com/sethgrid/pester"
)

var (
	padURL           = flag.String("p", "", "link to pad")
	gitRepo          = flag.String("g", "", "path to repo, with ssh syntax, and write access")
	name             = flag.String("n", "", "name of the file, automatically set to pad name")
	contentSizeLimit = flag.Int64("l", 10485760, "limit of content to fetch and commit")
)

// fetchPad fetches the text from the pad.
func fetchPad(link string) ([]byte, error) {
	resp, err := pester.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("got %s at %s", resp.Status, link)
	}
	return ioutil.ReadAll(io.LimitReader(resp.Body, *contentSizeLimit))
}

func main() {
	flag.Parse()

	if *padURL == "" || *gitRepo == "" {
		log.Fatal("Pad -p and git repo -g are required")
	}

	u, err := url.Parse(*padURL)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = path.Join(u.Path, "export/txt")
	log.Printf("export URL at: %s", u.String())

	data, err := fetchPad(u.String())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("fetched %d bytes from %s", len(data), *padURL)

	// Clone repo into temporary dir, add file, commit, push.
	dir, err := ioutil.TempDir("", "padsync-git-")
	if err != nil {
		log.Fatal(err)
	}
	err = clam.Run("git clone {{ repo }} {{ dir }}", clam.Map{"repo": *gitRepo, "dir": dir})
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dir)
	filename := fmt.Sprintf("%s.txt", filepath.Join(dir, slug.Make(*padURL)))
	log.Printf("updating %s", filename)
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Fatal(err)
	}
	if err := clam.Run(`cd {{ dir }} && git add {{ filename }} && git commit -m "auto-commit" && git push origin master && cd -`,
		clam.Map{"dir": dir, "filename": filename}); err != nil {
		log.Fatal(err)
	}
	log.Println("successfully updated repo")
}
