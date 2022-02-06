package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/joho/godotenv"
)

var (
	chunkSize uint64 = 1048576
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
	fc := files.New(*client)
	lfa := files.NewListFolderArg(path)
	lfr, err := fc.ListFolder(lfa)

	if err != nil {
		return nil, err
	}

	for _, v := range lfr.Entries {
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
	fc := files.New(*client)
	da := files.NewDownloadArg(from)
	_, src, err := fc.Download(da)

	if err != nil {
		return err
	}

	defer src.Close()
	file, err := os.Create(to)

	if err != nil {
		return err
	}

	defer file.Close()
	_, err = io.Copy(file, src)

	return
}

func GetRootDir() (rootDir string) {
	workingDir, err := os.Getwd()

	if err != nil {
		panic("Error to get working directory")
	}

	rootDir = filepath.Dir(workingDir)

	return
}

func LoadEnv() {
	err := godotenv.Load(path.Join(GetRootDir(), "../.env"))

	if err != nil {
		panic("Error to load environment variables")
	}
}

func main() {
	LoadEnv()

	client := NewClient(os.Getenv("DROPBOX_AUTH_TOKEN"))

	nodes, err := ListFilesAndFolders(&client, "")

	if err != nil {
		println("Error: ", err.Error())
	}

	for i, n := range nodes {
		println("Node: ", n.Name, ", Index: ", i)
	}
}
