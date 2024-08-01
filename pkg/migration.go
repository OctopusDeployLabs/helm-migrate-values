package pkg

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"regexp"
)

type Migration struct {
	From version.Version
	To   version.Version
}

// From https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string (named groups version)
var semVerRegEx = "(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?"

func NewMigration(fileName string) (*Migration, error) {

	// Get version string from file name, eg '1.0.0-1.0.1.yaml'
	pattern := fmt.Sprintf(`(?P<fromVersion>%s)\-(?P<toVersion>%s)\.yaml`, semVerRegEx, semVerRegEx)
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(fileName)
	names := re.SubexpNames()
	var from, to string
	for i, match := range matches {
		if names[i] == "fromVersion" {
			from = match
		} else if names[i] == "toVersion" {
			to = match
		}
	}

	fromVersion, err := version.NewVersion(from)
	if err != nil {
		return nil, fmt.Errorf("error parsing 'from' version '%s'': %v", from, err)
	}

	toVersion, err := version.NewVersion(to)
	if err != nil {
		return nil, fmt.Errorf("error parsing 'to' version '%s': %v", to, err)
	}

	if fromVersion.GreaterThanOrEqual(toVersion) {
		return nil, fmt.Errorf("migration 'from' versions must be less than their 'to' version")
	}

	return &Migration{
		From: *fromVersion,
		To:   *toVersion,
	}, nil
}
