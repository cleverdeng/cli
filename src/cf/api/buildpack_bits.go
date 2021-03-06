package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fileutils"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type BuildpackBitsRepository interface {
	UploadBuildpack(buildpack cf.Buildpack, dir string) (apiResponse net.ApiResponse)
}

type CloudControllerBuildpackBitsRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
	zipper  cf.Zipper
}

func NewCloudControllerBuildpackBitsRepository(config *configuration.Configuration, gateway net.Gateway, zipper cf.Zipper) (repo CloudControllerBuildpackBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerBuildpackBitsRepository) UploadBuildpack(buildpack cf.Buildpack, dir string) (apiResponse net.ApiResponse) {
	fileutils.TempFile("buildpack", func(zipFile *os.File, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithMessage(err.Error())
			return
		}
		err = repo.zipper.Zip(dir, zipFile)
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Invalid buildpack", err)
			return
		}
		apiResponse = repo.uploadBits(buildpack, zipFile)
		if apiResponse.IsNotSuccessful() {
			return
		}
	})
	return
}

func (repo CloudControllerBuildpackBitsRepository) uploadBits(app cf.Buildpack, zipFile *os.File) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.Target, app.Guid)

	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithMessage(err.Error())
			return
		}

		boundary, err := repo.writeUploadBody(zipFile, requestFile)
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error creating upload", err)
			return
		}

		request, apiResponse := repo.gateway.NewRequest("PUT", url, repo.config.AccessToken, requestFile)
		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
		request.HttpReq.Header.Set("Content-Type", contentType)
		if apiResponse.IsNotSuccessful() {
			return
		}

		apiResponse = repo.gateway.PerformRequest(request)
	})

	return
}

func (repo CloudControllerBuildpackBitsRepository) writeUploadBody(zipFile *os.File, body *os.File) (boundary string, err error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	zipStats, err := zipFile.Stat()
	if err != nil {
		return
	}

	if zipStats.Size() == 0 {
		return
	}

	part, err := writer.CreateFormFile("buildpack", "buildpack.zip")
	if err != nil {
		return
	}

	_, err = io.Copy(part, zipFile)
	return
}
