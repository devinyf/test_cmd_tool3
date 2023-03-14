package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type TempData struct {
	ModuleName  string
	PackageName string
	FuncName    string
}

const githubUrl = "git@github.com:dyfsquall/gin-kratos-temp.git"
const TEMPLATE_DIR = "./.testTargetDir"

func doGitClone(auth transport.AuthMethod) error {
	_, err := git.PlainClone(TEMPLATE_DIR,
		false,
		&git.CloneOptions{
			URL:               githubUrl,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Auth:              auth,
			// Progress:          os.Stdout,
		})

	return err
}

var (
	PROJECT_NAME string
)

func deleteGitClone() {
	os.RemoveAll(TEMPLATE_DIR)
}

func getPublicKey() *ssh.PublicKeys {
	// Username must be "git" for SSH auth to work, not your real username.
	// See https://github.com/src-d/go-git/issues/637
	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	sshKey, err := os.ReadFile(sshPath)
	if err != nil {
		panic(err)
	}

	publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if err != nil {
		log.Fatalf("generate publicKey failed")
	}

	return publicKey

}

func getHttpBasicAuth() *http.BasicAuth {
	// TODO: using config file
	return &http.BasicAuth{
		Username: "<username>",
		Password: "<password>",
	}
}

func main() {
	// err := doGitClone(getHttpBasicAuth())
	err := doGitClone(getPublicKey())
	if err != nil {
		panic(err)
	}

	defer deleteGitClone()

	pjName := flag.String("name", "", "new_project_name")
	flag.Parse()

	if *pjName == "" {
		fmt.Println("Must-Have-ProjectName use: --name xxx")
		return
	}

	PROJECT_NAME = *pjName

	// TODO ...
	GetAllFile(TEMPLATE_DIR)
}

func GetAllFile(pathname string) error {
	ParseFolder(pathname)

	// 深度遍历文件夹
	// rd, err := ioutil.ReadDir(pathname)
	rd, err := os.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			err = GetAllFile(fullDir)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return err
			}
		}
	}
	return nil
}

// 解析模版 创建文件
func ParseFolder(pathname string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic recoverd: err: %v\n", r)
		}
	}()

	dir := strings.Replace(pathname, TEMPLATE_DIR, PROJECT_NAME, 1)
	tryCreateDir(dir)

	temp, err := template.ParseGlob(pathname + "/*.temp")
	if err != nil {
		if !strings.Contains(err.Error(), "pattern matches no files") {
			fmt.Printf("temp ParseGlob Err: %v", err.Error())
		}
		fmt.Println(err.Error())
		return
	}

	data := TempData{ModuleName: PROJECT_NAME}

	for _, vtemp := range temp.Templates() {

		fileName := vtemp.Name()
		ext := path.Ext(fileName)
		if ext == ".temp" {
			fileName = strings.TrimSuffix(vtemp.Name(), ext)
		}

		fullPath := dir + "/" + fileName

		f, err := os.Create(fullPath)
		if err != nil {
			fmt.Println("os Create Err: ", err.Error())
			panic(err)
		}

		err = vtemp.Execute(f, data)
		if err != nil {
			fmt.Println("template Exec Err: ", err.Error())
			panic(err)
		}
	}
}

func tryCreateDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("checkdir: ", dir)
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			fmt.Println("os Mkdir Err: ", err.Error())
			panic(err)
		}
	}
}
