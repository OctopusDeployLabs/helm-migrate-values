package main

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"slices"
)

func getRelease(name string, listAction *action.List) (*release.Release, error) {
	listAction.Deployed = true
	listAction.SetStateMask()

	releases, err := listAction.Run()
	if err != nil {
		return nil, err
	}

	idx := slices.IndexFunc(releases, func(r *release.Release) bool { return r.Name == name })
	if idx == -1 {
		return nil, errors.New("Could not find a Helm release matching the given release name.")
	}

	return releases[idx], nil
}
