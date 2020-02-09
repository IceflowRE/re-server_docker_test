package updater

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func getFileHash(filename string) *string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	res := "sha256:" + hex.EncodeToString(hasher.Sum(nil))
	return &res
}

type baseResp struct {
	Images []image `json:"images"`
}

type image struct {
	Os     string `json:"os"`
	Arch   string `json:"architecture"`
	Digest string `json:"digest"`
}

func getAlpineHash(arch string, os string) *string {
	resp, err := http.Get("https://registry.hub.docker.com/v2/repositories/library/alpine/tags/latest")
	if err != nil {
		return nil
	}
	if resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var respUn baseResp
	err = json.Unmarshal(body, &respUn)
	for _, image := range respUn.Images {
		if image.Arch == arch && image.Os == os {
			return &image.Digest
		}
	}
	fmt.Println("base image for "+arch+", ", os+" not found")
	return nil
}

func getCommitHash(branch string) *string {
	out, err := exec.Command("git", "ls-remote", "https://github.com/redeclipse/base.git", "refs/heads/"+branch).Output()
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	res := "sha256:" + string(bytes.Split(out, []byte("\t"))[0])
	return &res
}

func getNewHashes(dockerfile string, branch string, arch string, os string) *hash {
	reCommit := getCommitHash(branch)
	if reCommit == nil {
		fmt.Println("failed to get newest git commit hash")
		return nil
	}
	dockerHash := getFileHash(dockerfile)
	if dockerHash == nil {
		fmt.Println("dockerfile hash failed " + dockerfile)
		return nil
	}
	alpineHash := getAlpineHash(arch, os)
	if alpineHash == nil {
		fmt.Println("alpine hash failed")
		return nil
	}
	return &hash{
		Alpine:     *alpineHash,
		Dockerfile: *dockerHash,
		ReCommit:   *reCommit,
	}
}
