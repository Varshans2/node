package ui

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/mysteriumnetwork/node/requests"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

const (
	whichFileName = "which.json"
)

type UIServer interface {
	SwitchUI(path string)
}

type VersionManager struct {
	uiServer   UIServer
	httpClient *requests.HTTPClient
	github     *github
	nodeUIDir  string
}

func NewVersionManager(uiServer UIServer, http *requests.HTTPClient, nodeUIDir string) *VersionManager {
	return &VersionManager{uiServer: uiServer, httpClient: http, nodeUIDir: nodeUIDir, github: newGithub(http)}
}

// TODO add logging
// TODO check integrity of downloaded release so not to serve a broken nodeUI
func (vm *VersionManager) ListLocal() ([]LocalVersion, error) {
	files, err := ioutil.ReadDir(vm.nodeUIDir)
	if err != nil {
		return nil, fmt.Errorf("could not read "+nodeUIPath+": %w", err)
	}

	var versions []LocalVersion
	versions = append(versions, LocalVersion{
		Version: Version{Name: "bundled"},
	})
	for _, f := range files {
		if f.IsDir() {
			versions = append(versions, LocalVersion{
				Version: Version{Name: f.Name()},
			})
		}
	}

	versionInUse, err := vm.Which()
	if err != nil {
		return nil, err
	}
	for i := range versions {
		if versions[i].Name == versionInUse {
			versions[i].InUse = true
		}
	}

	return versions, nil
}

// TODO since there are github api calls maybe have some cache so to try and avoid rate limit for unauth calls
// TODO pagination
func (vm *VersionManager) ListRemote() ([]Version, error) {
	releases, err := vm.github.nodeUIReleases()
	if err != nil {
		return nil, err
	}

	var versions []Version
	for _, release := range releases {
		versions = append(versions, Version{
			Name: release.Name,
		})
	}

	return versions, nil
}

// TODO avoid multiple downloads at the same time
// TODO introduce download progress
// TODO think about sending SSE to inform to nodeUI that a version has been downloaded

func (vm *VersionManager) Download(versionName string) error {
	urlString, err := vm.github.nodeUIDownloadURL(versionName)
	if err != nil {
		return err
	}

	assetURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("failed to parse asset download URL: %w", err)
	}

	err = vm.downloadTarball(assetURL, versionName)
	if err != nil {
		return err
	}

	return vm.untarAndExplode(versionName)
}

func (vm *VersionManager) downloadTarball(url *url.URL, versionName string) error {
	req, err := requests.NewGetRequest(fmt.Sprintf("%s://%s", url.Scheme, url.Host), url.RequestURI(), nil)
	if err != nil {
		return fmt.Errorf("failed to create NodeUI releases fetch request: %w", err)
	}

	resp, err := vm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request asset: %w", err)
	}
	defer resp.Body.Close()

	err = os.MkdirAll(vm.uiDistPath(versionName), 0700)
	if err != nil {
		return fmt.Errorf("failed to create path: %w", err)
	}

	out, err := os.Create(vm.uiDistFile(versionName))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download the asset: %w", err)
	}

	return nil
}

func (vm *VersionManager) SwitchTo(versionName string) error {
	local, err := vm.ListLocal()
	if err != nil {
		return err
	}
	for _, lv := range local {
		if lv.Version.Name == versionName {
			vm.writeWhichCFG(which{VersionName: versionName})
			vm.uiServer.SwitchUI(vm.uiBuildPath(versionName))
			return nil
		}
	}

	return errors.New("no local version named: " + versionName)
}

func (vm *VersionManager) Which() (string, error) {
	if !vm.whichCFGExists() {
		return "bundled", nil
	}

	w, err := vm.readWhichCFG()
	return w.VersionName, err
}

// TODO maybe can reduce number of paths

func (vm *VersionManager) whichFile() string {
	return fmt.Sprintf("%s/%s", vm.nodeUIDir, whichFileName)
}

func (vm *VersionManager) uiDistPath(versionName string) string {
	return fmt.Sprintf("%s/%s", vm.nodeUIDir, versionName)
}

func (vm *VersionManager) uiBuildPath(versionName string) string {
	return fmt.Sprintf("%s/%s/build", vm.nodeUIDir, versionName)
}

func (vm *VersionManager) uiDistFile(versionName string) string {
	return fmt.Sprintf("%s/%s", vm.uiDistPath(versionName), nodeUIAssetName)
}

func (vm *VersionManager) whichCFGExists() bool {
	if _, err := os.Stat(vm.whichFile()); os.IsNotExist(err) {
		return false
	}

	return true
}

func (vm *VersionManager) readWhichCFG() (which, error) {
	data, err := ioutil.ReadFile(vm.whichFile())
	if err != nil {
		return which{}, err
	}

	var w which
	err = json.Unmarshal(data, &w)
	if err != nil {
		return which{}, err
	}

	return w, nil
}

func (vm *VersionManager) writeWhichCFG(w which) {
	json, err := json.Marshal(w)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(vm.whichFile(), json, 644)
	return
}

func (vm *VersionManager) untarAndExplode(versionName string) error {
	file, err := os.Open(vm.uiDistFile(versionName))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	err = untar(vm.uiDistPath(versionName), file)
	if err != nil {
		return fmt.Errorf("failed to untar nodeUI dist: %w", err)
	}

	return nil
}

type LocalVersion struct {
	Version
	InUse bool `json:"in_use"`
}

type Version struct {
	Name string `json:"name"`
}

// TODO think about moving `which` logic elsewhere since VersionManager and Server will use it
type which struct {
	VersionName string `json:"version_name"`
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
// https://gist.githubusercontent.com/sdomino/635a5ed4f32c93aad131/raw/1f1a2609f9bf04f3a681a96c26350b0d694549bf/untargz.go
func untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
