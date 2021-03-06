package controllers

import (
	"errors"
	"fmt"
	"os/exec"
	"path"

	"github.com/devclub-iitd/DeployBot/src/helper"
	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// stop stops a running service based on the response from slack
func stop(callbackID string, data map[string]interface{}) {
	channel := data["channel"].(string)
	actionLog := history.NewAction("stop", data)
	if err := slack.PostChatMessage(channel, actionLog.String(), nil); err != nil {
		log.Warnf("error occured in posting chat message - %v", err)
		return
	}
	log.Infof("beginning %s with callback_id as %s", actionLog, callbackID)

	logPath := fmt.Sprintf("stop/%s.txt", callbackID)

	output, err := internalStop(actionLog)
	helper.WriteToFile(path.Join(logDir, logPath), string(output))
	actionLog.LogPath = logPath
	if err != nil {
		actionLog.Result = "failed"
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, fmt.Sprintf("%s\nError: %s\n", actionLog, err.Error()), nil)
	} else {
		actionLog.Result = "success"
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, actionLog.String(), nil)
	}
}

// internalStop actually runs the script to stop the given app.
func internalStop(a *history.ActionInstance) ([]byte, error) {
	state := history.GetState(a.RepoURL)

	var output []byte
	var err error

	switch state.Status {
	case "deploying":
		log.Infof("service(%s) is being deployed", a.RepoURL)
		output = []byte("Service is being deployed. Please wait for the process to be completed and try again.")
		err = errors.New("cannot stop while deploying")
	case "stopping":
		log.Infof("service(%s) is being stopped. Can't start another stop instance", a.RepoURL)
		output = []byte("Service is already being stopped.")
		err = errors.New("already stopping")
	case "running":
		log.Infof("calling %s to stop service(%s)", stopScriptName, a.RepoURL)
		state.Status = "stopping"
		history.SetState(a.RepoURL, state)
		if output, err = exec.Command(stopScriptName, state.Subdomain, a.RepoURL, state.Server).CombinedOutput(); err != nil {
			state.Status = "running"
			history.SetState(a.RepoURL, state)
		} else {
			state.Status = "stopped"
			history.SetState(a.RepoURL, state)
		}
	default:
		log.Infof("service(%s) is already stopped", a.RepoURL)
		output = []byte("Service is already stopped!")
		err = errors.New("already stopped")
	}
	return output, err
}
