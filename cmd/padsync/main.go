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
	dest             = flag.String("t", "", "destination path (or slug of url, if not specified)")
	branch           = flag.String("b", "main", "branch name")
)

// readBodyLimit fetches the text from the pad. Any HTTP error will be treated
// as an error as well.
func readBodyLimit(link string, limit int64) ([]byte, error) {
	resp, err := pester.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("got %s at %s", resp.Status, link)
	}
	return ioutil.ReadAll(io.LimitReader(resp.Body, limit))
}

func main() {
	var (
		u                          *url.URL
		data                       []byte
		err                        error
		tempDir, command, filename string
	)
	flag.Usage = func() {
		fmt.Println(`Usage: padsynpadsync [OPTIONS]

    $ padsync -g git@github.com/example/repo -p https://etherpad.wikimedia.org/p/padsync-example -b master

Flags
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *padURL == "" || *gitRepo == "" {
		log.Fatal("pad -p and git repo -g are required")
	}
	if u, err = url.Parse(*padURL); err != nil {
		log.Fatal(err)
	}
	u.Path = path.Join(u.Path, "export/txt")
	log.Printf("export URL at: %s", u.String())
	if data, err = readBodyLimit(u.String(), *contentSizeLimit); err != nil {
		log.Fatal(err)
	}
	log.Printf("fetched %d bytes from %s", len(data), *padURL)
	if tempDir, err = ioutil.TempDir("", "padsync-git-"); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("cache directory at %s", tempDir)
	command = fmt.Sprintf(`git clone "%s" "%s"`, *gitRepo, tempDir)
	switch {
	case *dryRun:
		log.Printf(command)
	default:
		if err = clam.Run(command, clam.Map{}); err != nil {
			log.Fatal(err)
		}
	}
	filename = fmt.Sprintf("%s.txt", filepath.Join(tempDir, slug.Make(*padURL)))
	if *dest != "" {
		filename = fmt.Sprintf(filepath.Join(tempDir, *dest))
		if err = os.MkdirAll(path.Dir(filename), 0755); err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("updating %s", filename)
	switch {
	case *dryRun:
		log.Printf("written file %s", filename)
	default:
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			log.Fatal(err)
		}
	}
	command = fmt.Sprintf(`
		cd "%s" &&
		git pull origin %s &&
		git add "%s" &&
		git diff-index --quiet HEAD || git commit -m "auto-commit" &&
		git push origin %s && cd -`,
		tempDir, *branch, filename, *branch)
	switch {
	case *dryRun:
		log.Println(command)
	default:
		if err := clam.Run(command, clam.Map{}); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("successfully updated repo")
}
