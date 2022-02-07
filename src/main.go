package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/joho/godotenv"
)

var (
	downloadsPath      = "../downloads"
	directorySeparator = filepath.FromSlash("/")
)

type Node struct {
	IsFolder       bool
	Name           string
	Size           uint64
	Rev            string
	ServerModified time.Time
}

func NewClient(token string) (client dropbox.Config) {
	client = dropbox.Config{
		Token:    token,
		LogLevel: dropbox.LogOff,
	}

	return
}

func ParseFileMetadata(file *files.FileMetadata) (n Node) {
	n.IsFolder = false
	n.Name = file.Name
	n.Size = file.Size
	n.Rev = file.Rev
	n.ServerModified = file.ServerModified

	return
}

func ParseFolderMetadata(folder *files.FolderMetadata) (n Node) {
	n.IsFolder = true
	n.Name = folder.Name

	return
}

func ListFilesAndFolders(client *dropbox.Config, path string) (nodes []Node, err error) {
	filesClient := files.New(*client)
	listFolderArg := files.NewListFolderArg(path)
	listFolderResult, err := filesClient.ListFolder(listFolderArg)

	if err != nil {
		return nil, err
	}

	for _, v := range listFolderResult.Entries {
		var n Node

		switch fm := v.(type) {
		case *files.FileMetadata:
			n = ParseFileMetadata(fm)
		case *files.FolderMetadata:
			n = ParseFolderMetadata(fm)
		}

		nodes = append(nodes, n)
	}

	return
}

func Download(client *dropbox.Config, from string, to string) (err error) {
	filesClient := files.New(*client)

	formattedFrom := strings.ReplaceAll(from, directorySeparator, "/")
	downloadArg := files.NewDownloadArg(formattedFrom)

	_, src, err := filesClient.Download(downloadArg)

	if err != nil {
		panic(err.Error())
	}

	folders := strings.Split(to, directorySeparator)
	folders = folders[:len(folders)-1]

	var pathWithoutFile string

	for _, folder := range folders {
		pathWithoutFile += directorySeparator + folder
	}

	err = os.MkdirAll(path.Join(GetRootDir(), pathWithoutFile), os.FileMode(0522))

	if err != nil {
		panic(err.Error())
	}

	to = path.Join(GetRootDir(), to)

	defer src.Close()
	file, err := os.Create(to)

	if err != nil {
		return err
	}

	defer file.Close()
	_, err = io.Copy(file, src)

	return
}

func WalkAndDownload(client *dropbox.Config, currentPath string, node Node) (err error) {
	if err == filepath.SkipDir {
		return nil
	}

	if node.IsFolder {
		nestedNodes, err := ListFilesAndFolders(client, currentPath)

		if err != nil {
			return err
		}

		if len(nestedNodes) == 0 {
			return filepath.SkipDir
		}

		for _, nestedNode := range nestedNodes {
			pathWithNestedNodeName := currentPath + "/" + nestedNode.Name

			println("WalkAndDownload call with ->", pathWithNestedNodeName)

			WalkAndDownload(client, pathWithNestedNodeName, nestedNode)
		}
	} else {
		println("Download call with ->", currentPath)

		formatedCurrentPath := strings.ReplaceAll(currentPath, "/", directorySeparator)

		err = Download(client, currentPath, path.Join(GetRootDir(), downloadsPath, formatedCurrentPath))

		if err != nil {
			println("Download error ->", err.Error())
		}
	}

	return
}

func GetRootDir() (rootDir string) {
	workingDir, err := os.Getwd()

	if err != nil {
		panic("Error to get working directory")
	}

	rootDir = workingDir

	return
}

func LoadEnv() {
	err := godotenv.Load(path.Join(GetRootDir(), ".env"))

	if err != nil {
		panic("Error to load environment variables")
	}
}

func main() {
	LoadEnv()

	client := NewClient(os.Getenv("DROPBOX_AUTH_TOKEN"))
	rootNodes, err := ListFilesAndFolders(&client, "")

	if err != nil {
		println("Error to list files and foldes")
	}

	for i, node := range rootNodes {
		println("Node ->", node.Name, "| Index ->", i)
	}

	println("-------------------------------------")

	exampleNode := rootNodes[0]
	exampleNodePath := "/" + exampleNode.Name

	WalkAndDownload(&client, exampleNodePath, exampleNode)
}
