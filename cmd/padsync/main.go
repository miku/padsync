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
	contentSizeLimit = flag.Int64("l", 10485760, "limit of content to fetch and commit")
	dryRun           = flag.Bool("dry", false, "dry run")
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

	log.Printf("cache directory at %s", dir)
	command := fmt.Sprintf(`git clone "%s" "%s"`, *gitRepo, dir)

	if *dryRun {
		log.Printf(command)
	} else {
		err = clam.Run(command, clam.Map{})
		if err != nil {
			log.Fatal(err)
		}
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Fatal(err)
		}
	}()
	filename := fmt.Sprintf("%s.txt", filepath.Join(dir, slug.Make(*padURL)))
	log.Printf("updating %s", filename)
	if !*dryRun {
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			log.Fatal(err)
		}
	}
	command = fmt.Sprintf(`cd "%s" && git pull && git add "%s" && git diff-index --quiet HEAD || git commit -m "auto-commit" && git push && cd -`, dir, filename)
	if *dryRun {
		log.Println(command)
	} else {
		if err := clam.Run(command, clam.Map{}); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("successfully updated repo")
}
