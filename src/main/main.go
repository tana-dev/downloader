package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Html struct {
	FileinfoList [][]string
	Breadcrumbs  map[string]string
	User         string
	Ip           string
	Bookmark     map[string]string
}

func handler(w http.ResponseWriter, r *http.Request) {

	var ip string
	var user string
	var url string
	var fileinfoList [][]string
	var breadcrumbs map[string]string
	var fpath string
	var fname string
	var bookmark map[string]string

	// 外出ししてもいいかも
	// ip = "10.27.145.100:8080" // Windows
	// ip = "10.27.148.99:8080" // kitahara-s
	ip = "10.27.144.136:8080" // kudo-mayu
	// ip = "192.168.33.22:8080" // Linux
	url = "http://" + ip + "/"
	user = "kudo-mayu"
	// user = "tanaka-shu"
	// user = "kitarara-s"
	bookmark = map[string]string{
		"グルメニュース":            url + "gn-fs11/pad/restaurant/00_share/【contents】/グルメニュース/",
		"EDMマーケティングオートメーション": url + "gn-fs11/pad/restaurant/00_share/【contents】/EDMマーケティングオートメーション/"}
	// "会員属性案件": url + "gn-fs11/pad/restaurant/00_share/【contents】/会員属性案件/",
	// "EDM進行管理": url + "gn-fs11/pad/restaurant/00_share/【contents】/EDM進行管理/"}

	fpath = r.URL.Path
	fpath1 := r.URL.Path
	fpath1 = strings.TrimRight(fpath1, "/")

	// pathを取るにはr.URL.Pathで受け取文末のスラッシュを削除
	fpath = `\` + strings.Replace(r.URL.Path, "/", `\`, -1) // 1.Windows
	fpath = strings.TrimRight(fpath, `\`)                   // 1.Windows
	// fpath = strings.TrimRight(fpath, "/") // 2. Linux
	fname = filepath.Base(fpath)

	// ファイル存在チェック
	fi, err := os.Stat(fpath)
	if err != nil {
		fmt.Fprintf(w, "ファイル、もしくはディレクトが存在しません")
		return
	}

	// breadcrumbs create
	dirs_list := strings.Split(strings.TrimLeft(fpath1, "/"), "/")
	breadcrumbs = map[string]string{}
	var indexs map[int]string
	indexs = map[int]string{}
	for i := 0; i < len(dirs_list); i++ {
		for l := 0; l <= i; l++ {
			if l == 0 {
				indexs[i] = dirs_list[l] + "/"
			} else {
				indexs[i] = indexs[i] + dirs_list[l] + "/"
			}
		}
		index := url + indexs[i]
		breadcrumbs[index] = dirs_list[i]
	}

	if fi.IsDir() {
		fpaths := dirwalk(fpath)
		for _, fp := range fpaths {
			var fileinfo []string
			var dir string
			link := strings.Replace(fp, `\`, "/", -1)      // 2.Windows
			link = url + strings.Replace(link, "/", "", 2) // 2.Windows
			// link := url + strings.Replace(fp, "/", "", 1) // 2.Linux
			name := filepath.Base(fp)
			f, _ := os.Stat(fp)
			if f.IsDir() {
				dir = "fa-folder"
			} else {
				dir = "fa-file-o"
			}

			if err != nil {
				fmt.Fprintf(w, "ファイルの読み込みに失敗しました")
				return
			}
			updatetime_tmp := f.ModTime()
			updatetime := updatetime_tmp.Format("2006-01-02 15:04:05")

			fileinfo = append(fileinfo, link)
			fileinfo = append(fileinfo, name)
			fileinfo = append(fileinfo, updatetime)
			fileinfo = append(fileinfo, dir)
			fileinfoList = append(fileinfoList, fileinfo)
		}
		// sort.Sort(fileinfoList)

	} else {
		ext := fname[strings.LastIndex(fname, "."):]
		out := readfile(fpath)
		ctype := createContentType(ext)
		w.Header().Set("Content-Disposition", "attachment; filename="+fname)
		w.Header().Set("Content-Type", ctype)
		// w.Header().Set("Content-Length", string(len(out)))
		w.Write(out)
		return
	}

	h := Html{
		FileinfoList: fileinfoList,
		Breadcrumbs:  breadcrumbs,
		User:         user,
		Ip:           ip,
		Bookmark:     bookmark,
	}

	// funcs := template.FuncMap{"add": add}
	// tmpl := template.Must(template.New("./view/index.html").Funcs(funcs).ParseFiles("./view/index.html"))
	// tmpl.Execute(w, h)

	templ_file, err := Asset("../resources/view/index.html")
	tmpl, _ := template.New("tmpl").Parse(string(templ_file))
	tmpl.Execute(w, h)

}

func main() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(assetFS())))
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func readfile(srcpath string) []byte {

	src, err := os.Open(srcpath)
	if err != nil {
		panic(err)
	}
	defer src.Close()

	contents, _ := ioutil.ReadAll(src)

	return contents
}

func copyfile(srcpath string, dstpath string) {

	src, err := os.Open(srcpath)
	if err != nil {
		panic(err)
	}
	defer src.Close()

	dst, err := os.Create(dstpath)
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
}

func createContentType(ext string) string {

	var ctype string

	// w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))

	switch ext {
	case ".txt":
		ctype = "text/plain"
	case ".csv":
		ctype = "text/csv" // CSVファイル
	case ".html":
		ctype = "text/html" // HTMLファイル
	case ".css":
		ctype = "text/css" // CSSファイル
	case ".js":
		ctype = "text/javascript" // JavaScriptファイル
	case ".exe":
		ctype = "application/octet-stream" // EXEファイルなどの実行ファイル
	case ".pdf":
		ctype = "application/pdf" // PDFファイル
	case ".xlsx":
		// ctype = "application/vnd.ms-excel" // EXCELファイル
		ctype = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" // EXCELファイル
	case ".ppt":
		ctype = "application/vnd.ms-powerpoint" // PowerPointファイル
	case ".docx":
		ctype = "application/msword" // WORDファイル
	case ".jpeg", ".jpg":
		ctype = "image/jpeg" // JPEGファイル(.jpg, .jpeg)
	case ".png":
		ctype = "image/png" // PNGファイル
	case ".gif":
		ctype = "image/gif" // GIFファイル
	case ".bmp":
		ctype = "image/bmp" // Bitmapファイル
	case ".zip":
		ctype = "application/zip" // Zipファイル
	case ".lzh":
		ctype = "application/x-lzh" // LZHファイル
	case ".tar":
		ctype = "application/x-tar" // tarファイル/tar&gzipファイル
	case ".mp3":
		ctype = "audio/mpeg" // MP3ファイル
	case ".mp4":
		ctype = "audio/mp4" // MP4ファイル
	case ".mpeg":
		ctype = "video/mpeg" // MPEGファイル（動画）
	default:
		ctype = "text/plain"
	}

	return ctype
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	var dpaths []string
	var fpaths []string
	for _, file := range files {
		if 0 != strings.Index(file.Name(), ".") && 0 != strings.Index(file.Name(), "~$") && 0 != strings.Index(file.Name(), "Thumbs.db") {

			f := filepath.Join(dir, file.Name())

			// ファイル存在チェック
			fi, _ := os.Stat(f)
			if fi.IsDir() {
				dpaths = append(dpaths, filepath.Join(dir, file.Name()))
			} else {
				fpaths = append(fpaths, filepath.Join(dir, file.Name()))
			}
		}
	}

	if nil == dpaths && nil != fpaths {
		paths = fpaths
	} else if nil != dpaths && nil == fpaths {
		paths = dpaths
	} else {
		paths = append(dpaths, fpaths...)
	}

	return paths
}

func add(x, y int) int {
	return x + y
}