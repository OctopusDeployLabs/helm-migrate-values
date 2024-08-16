package pkg

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"strings"
)

const FilePrefix = "to-v"

// const ExcludeFilePrefix = "exclude-to-v"
const FileExtension = ".yaml"

func parseFilenameIntoMigration(filename string) (*Migration, error) {

	// Get version string from file name, eg 'to-v1.0.1.yaml'
	versionString := strings.TrimPrefix(filename, FilePrefix)
	versionString = strings.TrimSuffix(versionString, FileExtension)

	version, err := version.NewVersion(versionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing version '%s': %v", versionString, err)
	}

	return &Migration{
		*version,
		filename,
	}, nil
}
