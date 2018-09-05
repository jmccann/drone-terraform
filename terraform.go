package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type (
	// Terraform holds input parameters for terraform
	Terraform struct {
		Version string
	}
)

func installExtraPem(pemName string, pemContents string) error {
	os.Mkdir(os.Getenv("HOME")+"/.ssh", 0700)
	err := ioutil.WriteFile(os.Getenv("HOME")+"/.ssh/"+pemName, []byte(pemContents), 0600)
	if err != nil {
		return err
	}
	return nil
}

func installGithubSsh(githubSshPrivate string) error {
	os.Mkdir(os.Getenv("HOME")+"/.aws", 0700)
	myconf := []byte("Host github.com\n    StrictHostKeyChecking no\n    UserKnownHostsFile=/dev/null\n")
	err := ioutil.WriteFile(os.Getenv("HOME")+"/.ssh/conf", myconf, 0644)
	if err != nil {
		return err
	}
	mykey := []byte(githubSshPrivate)
	err2 := ioutil.WriteFile(os.Getenv("HOME")+"/.ssh/id_rsa", mykey, 0600)
	if err2 != nil {
		return err2
	}
	return nil
}

func installProfile(profileName string, profileKey string, profileSecret string) error {
	os.Mkdir(os.Getenv("HOME")+"/.aws", 0700)
	myconf := []byte("[" + profileName + "]\naws_access_key_id = " + profileKey + "\naws_secret_access_key = " + profileSecret + "\n")
	err := ioutil.WriteFile(os.Getenv("HOME")+"/.aws/credentials", myconf, 0644)
	return err
}

func installTerraform(version string) error {
	err := downloadTerraform(version)
	if err != nil {
		return nil
	}

	return Unzip("/var/tmp/terraform.zip", "/bin")
}

func downloadTerraform(version string) error {
	return downloadFile("/var/tmp/terraform.zip", fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_linux_amd64.zip", version, version))
}

func downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Unzip a file to a destination
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
