/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"errors"
	"fmt"
)

var (
	ErrApplicationExists = errors.New("The application already exists in marathon, you must update")
)

type Applications struct {
	Apps []Application `json:"apps"`
}

type ApplicationWrap struct {
	Application Application `json:"app"`
}

type Application struct {
	ID            string            `json:"id",omitempty`
	Cmd           string            `json:"cmd,omitempty"`
	Constraints   [][]string        `json:"constraints,omitempty"`
	Container     *Container        `json:"container,omitempty"`
	CPUs          float32           `json:"cpus,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Executor      string            `json:"executor,omitempty"`
	HealthChecks  []*HealthCheck    `json:"healthChecks,omitempty"`
	Instances     int               `json:"instances,omitemptys"`
	Mem           float32           `json:"mem,omitempty"`
	Tasks         []*Task           `json:"tasks,omitempty"`
	Ports         []int             `json:"ports,omitempty"`
	RequirePorts  bool              `json:"requirePorts,omitempty"`
	BackoffFactor float32           `json:"backoffFactor,omitempty"`
	TasksRunning  int               `json:"tasksRunning,omitempty"`
	TasksStaged   int               `json:"tasksStaged,omitempty"`
	Uris          []string          `json:"uris,omitempty"`
	Version       string            `json:"version,omitempty"`
}

type ApplicationVersions struct {
	Versions []string `json:"versions"`
}

type ApplicationVersion struct {
	Version string `json:"version"`
}

func (client *Client) Applications() (*Applications, error) {
	applications := new(Applications)
	if err := client.ApiGet(MARATHON_API_APPS, "", applications); err != nil {
		return nil, err
	} else {
		return applications, nil
	}
}

func (client *Client) ListApplications() ([]string, error) {
	if applications, err := client.Applications(); err != nil {
		return nil, err
	} else {
		list := make([]string, 0)
		for _, application := range applications.Apps {
			list = append(list, application.ID)
		}
		return list, nil
	}
}

func (client *Client) HasApplicationVersion(name, version string) (bool, error) {
	if versions, err := client.ApplicationVersions(name); err != nil {
		return false, err
	} else {
		if Contains(versions.Versions, version) {
			return true, nil
		}
		return false, nil
	}
}

func (client *Client) ApplicationVersions(name string) (*ApplicationVersions, error) {
	uri := fmt.Sprintf("%s%s/versions", MARATHON_API_APPS, name)
	versions := new(ApplicationVersions)
	if err := client.ApiGet(uri, "", versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func (client *Client) ChangeApplicationVersion(name string, version *ApplicationVersion) (*DeploymentID, error) {
	client.Debug("Changing the application: %s to version: %s", name, version)
	uri := fmt.Sprintf("%s%s", MARATHON_API_APPS, name)
	deploymentId := new(DeploymentID)
	if err := client.ApiPut(uri, version, deploymentId); err != nil {
		client.Debug("Failed to change the application to version: %s, error: %s", version.Version, err)
		return nil, err
	}
	return deploymentId, nil
}

func (client *Client) Application(id string) (*Application, error) {
	application := new(ApplicationWrap)
	if err := client.ApiGet(fmt.Sprintf("%s%s", MARATHON_API_APPS, id), "", application); err != nil {
		return nil, err
	} else {
		return &application.Application, nil
	}
}

func (client *Client) ApplicationOK(name string) (bool, error) {
	/* step: check the application even exists */
	if found, err := client.HasApplication(name); err != nil {
		return false, err
	} else if !found {
		return false, ErrDoesNotExist
	}
	/* step: get the application */
	if application, err := client.Application(name); err != nil {
		return false, err
	} else {
		/* step: if the application has not health checks, just return true */
		if application.HealthChecks == nil || len(application.HealthChecks) <= 0 {
			return true, nil
		}
		/* step: does the application have any tasks */
		if application.Tasks == nil || len(application.Tasks) <= 0 {
			return true, nil
		}

		/* step: iterate the application checks and look for false */
		for _, task := range application.Tasks {
			if task.HealthCheckResult != nil {
				for _, check := range task.HealthCheckResult {
					if !check.Alive {
						return false, nil
					}
				}

			}
		}
		return true, nil
	}
}

func (client *Client) CreateApplication(application *Application) (bool, error) {
	/* step: check of the application already exists */
	if found, err := client.HasApplication(application.ID); err != nil {
		return false, err
	} else if found {
		return false, ErrApplicationExists
	}
	/* step: post the application to marathon */
	if err := client.ApiPost(MARATHON_API_APPS, &application, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (client *Client) HasApplication(name string) (bool, error) {
	client.Debug("Checking if application: %s exists in marathon", name)
	if name == "" {
		return false, ErrInvalidArgument
	} else {
		if applications, err := client.ListApplications(); err != nil {
			return false, err
		} else {
			for _, id := range applications {
				if name == id {
					client.Debug("The application: %s presently exist in maration", name)
					return true, nil
				}
			}
		}
		return false, nil
	}
}

func (client *Client) DeleteApplication(application *Application) (bool, error) {
	/* step: check of the application already exists */
	if found, err := client.HasApplication(application.ID); err != nil {
		return false, err
	} else if found {
		return false, ErrDoesNotExist
	} else {
		/* step: delete the application */
		client.Debug("Deleting the application: %s", application.ID)
		if err := client.ApiDelete(fmt.Sprintf("%s%s", MARATHON_API_APPS, application.ID), "", nil); err != nil {
			return false, err
		} else {

		}
	}
	return false, nil
}

func (client *Client) RestartApplication(application *Application, force bool) (*Deployment, error) {
	client.Debug("Restarting the application: %s, force: %s", application, force)
	/* step: check the application exists to restart */
	if found, err := client.HasApplication(application.ID); err != nil {
		return nil, err
	} else if found {
		return nil, ErrApplicationExists
	}
	return nil, nil
}

func (client *Client) ScaleApplication(application *Application, instances int) (*Deployment, error) {
	client.Debug("ScaleApplication: application: %s, instance: %d", application, instances)
	deployment := new(Deployment)
	if found, err := client.HasApplication(application.ID); err != nil {
		return nil, err
	} else if !found {
		return nil, ErrDoesNotExist
	}
	return deployment, nil
}
