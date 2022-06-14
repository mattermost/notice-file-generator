package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func CreateNoticeDir(config *Config) error {
	if _, err := os.Stat(config.NoticeWorkPath()); os.IsExist(err) {
		os.RemoveAll(config.NoticeWorkPath())
	}

	err := os.MkdirAll(config.NoticeWorkPath(), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func UpdateNotice(config *Config, dependencies []Dependency) error {
	out, err := os.OpenFile(config.NoticeFilePath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	writer := bufio.NewWriter(out)

	if _, err = writer.WriteString(fmt.Sprintf("%s\n\n", config.Title)); err != nil {
		log.Printf("Error while writing string %v", err)
	}
	if _, err = writer.WriteString(fmt.Sprintf("%s\n\n", config.Copyright)); err != nil {
		log.Printf("Error while writing string %v", err)
	}
	if _, err = writer.WriteString("NOTICES:\n--------\n\n"); err != nil {
		log.Printf("Error while writing string %v", err)
	}
	if _, err = writer.WriteString(fmt.Sprintf("%s\n\n", config.Description)); err != nil {
		log.Printf("Error while writing string %v", err)
	}
	if _, err = writer.WriteString("--------\n\n"); err != nil {
		log.Printf("Error while writing string %v", err)
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})

	idx := 0
	for _, d := range dependencies {
		if idx > 0 {
			if _, err = writer.WriteString("---\n\n"); err != nil {
				log.Printf("Error while writing string %v", err)
			}
		}
		if _, err = writer.WriteString(d.Load(config)); err != nil {
			log.Printf("Error while writing string %v", err)
		}
		idx = idx + 1
	}
	writer.Flush()
	return nil
}

func SplitExistingNotice(config *Config) error {
	noticeDir := config.NoticeDirPath()
	if _, err := os.Stat(noticeDir); os.IsNotExist(err) {
		err = os.MkdirAll(noticeDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if file, err := os.Open(config.NoticeFilePath()); err == nil {

		defer file.Close()

		scanner := bufio.NewScanner(file)
		var out *os.File
		var writer *bufio.Writer
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "## ") {
				if writer != nil {
					writer.Flush()
					out.Close()
					writer = nil
					out = nil
				}
				name := strings.Replace(line, "## ", "", -1)
				log.Printf("Found %s in existing notice.txt", name)
				out, err = os.OpenFile(noticeDir+"/"+GenerateFileName(name), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
				writer = bufio.NewWriter(out)
			}
			if out != nil {
				if strings.HasPrefix(line, "---") {
					writer.Flush()
					out.Close()
					writer = nil
					out = nil
				} else {
					if _, err = writer.WriteString(line + "\n"); err != nil {
						log.Printf("Error while writing string %v", err)
					}
				}
			}
		}
		if writer != nil {
			writer.Flush()
			out.Close()
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return nil

}

func MoveExistingNotice(config *Config, filename string) error {
	oldLocation := fmt.Sprintf("%s/%s", config.NoticeDirPath(), filename)
	newLocation := fmt.Sprintf("%s/%s", config.NoticeWorkPath(), filename)
	err := os.Rename(oldLocation, newLocation)
	if err != nil {
		return err
	}
	return nil
}
