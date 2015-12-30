package cpanel

import "github.com/letsencrypt-cpanel/cpanelgo"

type MkdirApiResponse struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		Permissions string `json:"permissions"`
		Name        string `json:"name"`
		Path        string `json:"path"`
	} `json:"data"`
}

func (c LiveApi) Mkdir(name, permissions, path string) (MkdirApiResponse, error) {
	var out MkdirApiResponse
	err := c.Gateway.API2("Fileman", "mkdir", cpanelgo.Args{
		"path":        path,
		"permissions": permissions,
		"name":        name,
	}, &out)

	return out, err
}

type UploadFilesApiResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Uploads   []string `json:"uploads"`
		Succeeded int      `json:"succeeded"`
		Warned    int      `json:"warned"`
		Failed    int      `json:"failed"`
	} `json:"data"`
}

func (c LiveApi) UploadFiles(name, contents, dir string) error {
	var out UploadFilesApiResponse
	err := c.Gateway.UAPI("Fileman", "upload_files", cpanelgo.Args{
		"dir":         dir,
		"name":        name,
		"contents":    contents,
		"letsencrypt": 1,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return err
}
