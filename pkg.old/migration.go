package pkg_old

import (
	"github.com/hashicorp/go-version"
)

type Migration struct {
	ToVersion version.Version
	FileName  string
}
