package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dchest/cssmin"
	"github.com/dchest/jsmin"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

func main() {
	source, _ := os.Getwd()
	output := path.Join(source, "compiled")
	version := strconv.FormatInt(time.Now().Unix(), 10)
	fmt.Println("Source Directory")
	fmt.Println("   ", source)
	fmt.Println("Output Directory")
	fmt.Println("    ", output)
	copyDir(source, output)

	directory, _ := os.Open(output)
	defer directory.Close()

	files, _ := directory.Readdir(-1)
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".html") {
			continue
		}

		fmt.Println("Parsing", f.Name())
		full := path.Join(directory.Name(), f.Name())
		handle, _ := os.Open(full)

		doc, _ := goquery.NewDocumentFromReader(handle)

		{
			fmt.Println("    Parsing Styles...")
			var css bytes.Buffer
			doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
				href, exists := s.Attr("href")
				if !exists {
					return
				}
				fmt.Println("        ->", href)
				if strings.HasPrefix(href, "http") {
					return
				}
				contents, _ := ioutil.ReadFile(path.Join(directory.Name(), href))
				css.WriteString(string(contents))
				css.WriteString("\n")
				s.Remove()
			})
			mini := cssmin.Minify(css.Bytes())
			o := "/css/" + f.Name() + "." + version + ".css"
			ioutil.WriteFile(path.Join(output, o), mini, f.Mode())
			doc.Find("head").AppendHtml("<link href='" + o + "' rel='stylesheet'/>")
		}

		{
			fmt.Println("    Parsing Scripts...")
			var js bytes.Buffer
			doc.Find("script").Each(func(i int, s *goquery.Selection) {
				if _, exists := s.Attr("debug"); exists {
					s.Remove()
					return
				}
				href, exists := s.Attr("src")
				if !exists {
					return
				}
				if strings.HasPrefix(href, "http") {
					return
				}
				absolute := path.Join(directory.Name(), href)
				var contents []byte
				if t, exists := s.Attr("type"); exists && t == "text/jsx" {
					fmt.Println("        ->", href, "(JSX)")
					c := exec.Command("jsx", absolute)
					contents, _ = c.Output()
				} else {
					fmt.Println("        ->", href)
					contents, _ = ioutil.ReadFile(absolute)
				}
				js.WriteString(string(contents))
				js.WriteString("\n")
				s.Remove()
			})
			mini, _ := jsmin.Minify(js.Bytes())
			//mini = js.Bytes()
			o := "/js/" + f.Name() + "." + version + ".js"
			ioutil.WriteFile(path.Join(output, o), mini, f.Mode())
			doc.Find("body").AppendHtml("<script src='" + o + "' />")
		}

		handle.Close()
		html, _ := doc.Html()
		ioutil.WriteFile(full, []byte(html), f.Mode())
	}

}
