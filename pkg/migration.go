package pkg

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"regexp"
)

type Migration struct {
	from     version.Version
	to       version.Version
	fileName string
}

// From https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string (non-named groups version)
// var semVerRegEx = "(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?"
var semVerRegEx = "(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?"

func NewMigration(fileName string) (*Migration, error) {

	// Get version string from file name, eg migration-v1.0.0-v1.0.1.yaml
	pattern := fmt.Sprintf(`migration-v(?P<fromVersion>%s)\-v(?P<toVersion>%s)\.yaml`, semVerRegEx, semVerRegEx)
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
		return nil, fmt.Errorf("error parsing 'from' version '%v'': %e", from, err)
	}

	toVersion, err := version.NewVersion(to)
	if err != nil {
		return nil, fmt.Errorf("error parsing 'to' version '%v': %e", to, err)
	}

	if fromVersion.GreaterThanOrEqual(toVersion) {
		return nil, fmt.Errorf("migration 'from; versions must be less than their 'to' version")
	}

	return &Migration{
		fileName: fileName,
		from:     *fromVersion,
		to:       *toVersion,
	}, nil
}

func EnsurePathExists(migrations []Migration, fromVer *version.Version, toVer *version.Version) error {

	fromVerExists, toVerExists := false, false

	for i, current := range migrations[:len(migrations)-1] {
		fromVerExists = fromVerExists || current.from.Equal(fromVer)
		toVerExists = toVerExists || current.to.Equal(toVer)

		next := migrations[i+1]

		if !current.to.Equal(&next.from) {
			return fmt.Errorf("migrations path is broken")
		}
	}

	if !fromVerExists || !toVerExists {
		return fmt.Errorf("no path between versions found")
	}

	return nil
}
