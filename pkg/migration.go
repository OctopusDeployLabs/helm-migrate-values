package pkg

import (
	"github.com/hashicorp/go-version"
)

type Migration struct {
	From version.Version
	To   version.Version
}
