package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const addonsPath = "C:/Users/Serg/Documents/Elder Scrolls Online/live/AddOns"
const esouiPath = "https://www.esoui.com/downloads/info"

func main() {
	start := time.Now()
	var count = 0
	fmt.Println("\033[33mНажмите любую клавишу для продолжения...\033[0m")
	fmt.Scanln()

	dir, err := os.Open(addonsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	fileInfos, err := dir.ReadDir(0)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {

			count++

			dirName := fileInfo.Name()
			fileName := dirName + ".txt"
			filePath := filepath.Join(addonsPath, dirName, fileName)

			file, err := os.Open(filePath)
			if err != nil {
				fmt.Println()
				log.Println(err)
				continue
			}
			fmt.Println()

			defer file.Close()

			fileTime, err := os.Stat(filePath)
			if err != nil {
				log.Fatal(err)
			}
			y, m, d := fileTime.ModTime().Date()

			addonDate := fmt.Sprintf("%02d/%02d/%d", m, d, y%100)
			fmt.Println("Открыт файл:", fileName+",", addonDate)
			addonVer := ""

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "## Title:") {
					fmt.Println("Title:", line[9:])
				}
				if strings.HasPrefix(line, "## Version:") {
					fmt.Println("Version:", line[11:])
					addonVer = line[12:]
					break
				} else if strings.HasPrefix(line, "## AddOnVersion:") {
					fmt.Println("AddOnVersion:", line[16:])
				}
			}

			if err := scanner.Err(); err != nil {
				log.Println(err)
			}

			fileData, err := os.Open("data.csv")

			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}

			defer fileData.Close()

			reader := csv.NewReader(fileData)
			reader.Comma = ';'

			records, err := reader.ReadAll()

			if err != nil {
				fmt.Println("Error reading CSV:", err)
				return
			}

			// поиск id аддона в базе для создания запроса
			upd := ""
			for _, record := range records {
				if record[0] == dirName {
					if record[1] != "" {
						upd = record[1]
						fmt.Println("id:", record[1])
					} else {
						fmt.Println("id:", "Не внесен в базу")
					}
				}
			}

			// get запрос
			if upd != "" {
				response, err := http.Get(esouiPath + upd)
				if err != nil {
					fmt.Printf("Не удалось выполнить GET-запрос: %s", err)
					return
				}
				defer response.Body.Close()

				// Инициализируем goquery для парсинга страницы
				doc, err := goquery.NewDocumentFromReader(response.Body)
				if err != nil {
					log.Fatalf("Не удалось создать документ goquery: %s", err)
				}

				// Находим тег <div id="safe">
				doc.Find("div#safe").Each(func(i int, s *goquery.Selection) {
					// Получаем текст, содержащийся в теге <div id="safe">
					text := s.Text()
					words := strings.Fields(text)
					//fmt.Println("Текст внутри тега <div id=\"safe\">:", words[1])
					//Сравниваем даты
					if words[1] == addonDate {
						fmt.Println("\033[32mUpdated!!!\033[0m")
					} else {
						//если даты разные пытаемся сравнить версии
						// Находим тег <div id="version">
						doc.Find("div#version").Each(func(i int, s *goquery.Selection) {
							// Получаем текст, содержащийся в теге <div id="version">
							text := s.Text()
							words := strings.Fields(text)
							if words[1] == addonVer {
								fmt.Println("\033[32mUpdated!!!\033[0m")
							} else {
								fmt.Println("\033[31mNeed update!\033[0m", esouiPath+upd, addonVer, "->", words[1])
							}
						})
					}
				})
			}
		}
	}
	fmt.Println()
	fmt.Println("\033[33mОбработано:", count, "\033[0m")
	elapsed := time.Since(start)
	fmt.Println("\033[33mВремя выполнения:", elapsed, "\033[0m")
}

// go build -ldflags="-s -w" main.go
