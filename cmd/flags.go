package cmd

import (
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

// This is the same flags as https://github.com/helm/helm/blob/main/cmd/helm/flags.go#L54
func addChartPathOptionsFlags(f *pflag.FlagSet, c *action.ChartPathOptions) {
	f.StringVar(&c.Version, "version", "", "specify a version constraint for the chart version to use. This constraint can be a specific tag (e.g. 1.1.1) or it may reference a valid range (e.g. ^2.0.0). If this is not specified, the latest version is used")
	f.BoolVar(&c.Verify, "verify", false, "verify the package before using it")
	f.StringVar(&c.Keyring, "keyring", defaultKeyring(), "location of public keys used for verification")
	f.StringVar(&c.RepoURL, "repo", "", "chart repository url where to locate the requested chart")
	f.StringVar(&c.Username, "username", "", "chart repository username where to locate the requested chart")
	f.StringVar(&c.Password, "password", "", "chart repository password where to locate the requested chart")
	f.StringVar(&c.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&c.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.BoolVar(&c.InsecureSkipTLSverify, "insecure-skip-tls-verify", false, "skip tls certificate checks for the chart download")
	f.BoolVar(&c.PlainHTTP, "plain-http", false, "use insecure HTTP connections for the chart download")
	f.StringVar(&c.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&c.PassCredentialsAll, "pass-credentials", false, "pass credentials to all domains")
}

// defaultKeyring returns the expanded path to the default keyring.
func defaultKeyring() string {
	if v, ok := os.LookupEnv("GNUPGHOME"); ok {
		return filepath.Join(v, "pubring.gpg")
	}
	return filepath.Join(homedir.HomeDir(), ".gnupg", "pubring.gpg")
}
