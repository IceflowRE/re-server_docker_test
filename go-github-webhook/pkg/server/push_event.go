package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/IceflowRE/redeclipse-server-docker/pkg/server/utils"
	"github.com/IceflowRE/redeclipse-server-docker/pkg/structs"
	"github.com/IceflowRE/redeclipse-server-docker/pkg/updater"
	"github.com/google/go-github/github"
)

var updateManager = utils.NewUpdateCtx()

func bindPushEvent(req *http.Request) (*github.PushEvent, *utils.ErrResp) {
	body, err := ioutil.ReadAll(req.Body)
	if utils.PrintError(err) {
		return nil, utils.GenericErrorResp
	}

	var obj github.PushEvent
	err = json.Unmarshal(body, &obj)
	if utils.PrintError(err) {
		return nil, &utils.ErrResp{Status: http.StatusBadRequest, Message: "Malformed body."}
	}
	switch {
	case obj.Ref == nil:
		return nil, &utils.ErrResp{Status: http.StatusBadRequest, Message: "Ref is empty."}
	}
	return &obj, nil
}

func pushEvent(hrw http.ResponseWriter, req *http.Request, updaterConfig *updater.Config, storage *structs.HashStorage, workDir string) {
	payload, err := bindPushEvent(req)
	if err != nil {
		utils.SendErrorResponse(hrw, err)
		return
	}
	githubHeader := req.Context().Value("header").(*utils.GithubHeader)

	buildConfig := updaterConfig.Get(*payload.Ref)
	if buildConfig == nil {
		utils.ResponseJSON(hrw, http.StatusOK,
			postResp{"Payload was not for an accepted reference, aborting.", githubHeader.Guid},
		)
		return
	}
	utils.ResponseJSON(hrw, http.StatusCreated,
		postResp{"Update started for '" + *payload.Ref + "'", githubHeader.Guid},
	)

	if start := updateManager.Add(*payload.Ref); !start {
		return
	}
	go func() {
		for {
			updater.Build(updaterConfig, storage, buildConfig, workDir)
			if start := updateManager.Remove(*payload.Ref); !start {
				break
			}
		}
	}()
	return
}
