package api

import (
	"cf"
	"cf/net"
)

type FakeAppFilesRepo struct{
	Application cf.Application
	Path string
	FileList string
}


func (repo *FakeAppFilesRepo)ListFiles(app cf.Application, path string) (files string, apiResponse net.ApiResponse) {
	repo.Application = app
	repo.Path = path

	files = repo.FileList

	return
}
