package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
)

func IndexOf[T comparable](collection []T, el T) int {
	for i, x := range collection {
		if x == el {
			return i
		}
	}
	return -1
}

func GenerateFileName(name string) string {
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")
	return reg.ReplaceAllString(name, "-")
}

func HTTPGet(rsc string) (string, error) {
	out := &bytes.Buffer{}

	client := http.Client{}

	req, err := http.NewRequest("GET", rsc, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("http status code %d when downloading %q", resp.StatusCode, rsc)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func main() {
	config := newConfig()

	log.Printf("Processing repo %s", config.Name)
	var err error
	var dependencies []Dependency

	if err = SplitExistingNotice(config); err != nil {
		log.Printf("Error occured while splitting existing notice.txt %s:%v", config.Name, err)
	}

	if dependencies, err = PopulateDependencies(config); err != nil {
		log.Fatalf("Error occured while populating dependencies %s:%v", config.Name, err)
	}
	if err = CreateNoticeDir(config); err != nil {
		log.Fatalf("Error occured while creating work folder %s:%v", config.Name, err)
	}

	var wg sync.WaitGroup

	for _, d := range dependencies {
		wg.Add(1)
		d := d
		go func() {
			defer wg.Done()

			if err = d.Generate(config); err != nil {
				log.Printf("Error occured while generating notice.txt %s:%v", d.Name, err)
			}

		}()
	}

	if err = UpdateNotice(config, dependencies); err != nil {
		log.Fatalf("Error occured while updating notice.txt %s:%v", config.Name, err)
	}
}
