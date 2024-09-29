package main

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var fileLine = regexp.MustCompile(`(?m)^\x{FEFF}?FILE "(.+)" .+$`)

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	var ok, ng, total int

	if len(os.Args) < 2 {
		fmt.Println("Usage: go-cue-fix [DIR]")
		os.Exit(1)
	}

	err := filepath.Walk(os.Args[1], func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ext := strings.ToLower(filepath.Ext(path)); ext != ".cue" {
			return nil
		}

		total++

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		buff, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		matches := fileLine.FindStringSubmatch(string(buff))
		if len(matches) != 2 {
			ng++
			log.Printf(`Can't find "FILE" line: "%s"`, path)
			return nil
		}

		realpath := filepath.Join(filepath.Dir(path), matches[1])
		if exists(realpath) {
			ok++
			return nil
		}

		utf8path, err := io.ReadAll(transform.NewReader(strings.NewReader(matches[1]), japanese.ShiftJIS.NewDecoder()))
		if err != nil {
			return err
		}

		realpath = filepath.Join(filepath.Dir(path), string(utf8path))
		if !exists(realpath) {
			ng++
			log.Printf(`File not exists: "%s" -> "%s"`, path, realpath)
			return nil
		}

		log.Printf("File exists (Convert ShiftJIS to UTF8): %s", realpath)

		result := strings.Replace(string(buff), matches[1], string(utf8path), 1)

		if err := os.WriteFile(path, []byte(result), 0644); err != nil {
			return err
		}

		ok++
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Total: %d, OK: %d, NG: %d", total, ok, ng)
}
