package build_info

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	// Injected by build_info script
	Version = "unknown-dev"
	Host    = "unknown" // Set by build_info script
	User    = "unknown" // Set by build_info script
	Stamp   = "0"       // Set by build_info script

	// Set by init()
	Date        time.Time
	IsRelease   bool
	IsCandidate bool
	IsBeta      bool
	LongVersion string

	// Set by Go build_info tags
	Tags []string

	allowedVersionExp = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[a-z0-9]+)*(\.\d+)*(\+\d+-g[0-9a-f]+)?(-[^\s]+)?$`)
)

func init() {
	if Version != "unknown-dev" {
		// If not a generic dev build_info, version string should come from git describe
		if !allowedVersionExp.MatchString(Version) {
			log.Fatalf("Invalid version string %q;\n\tdoes not match regexp %v", Version, allowedVersionExp)
		}
	}
	setBuildData()
}

func setBuildData() {
	// Check for a clean release build_info. A release is something like
	// "v0.1.2", with an optional suffix of letters and dot separated
	// numbers like "-beta3.47". If there's more stuff, like a plus sign and
	// a commit hash and so on, then it's not a release. If it has a dash in
	// it, it's some sort of beta, release candidate or special build_info. If it
	// has "-rc." in it, like "v0.14.35-rc.42", then it's a candidate build_info.
	//
	// So, every build_info that is not a stable release build_info has IsBeta = true.
	// This is used to enable some extra debugging (the deadlock detector).
	//
	// Release candidate builds are also "betas" from this point of view and
	// will have that debugging enabled. In addition, some features are
	// forced for release candidates - auto upgrade, and usage reporting.

	exp := regexp.MustCompile(`^v\d+\.\d+\.\d+(-[a-z]+[\d\.]+)?$`)
	IsRelease = exp.MatchString(Version)
	IsCandidate = strings.Contains(Version, "-rc.")
	IsBeta = strings.Contains(Version, "-")

	stamp, _ := strconv.Atoi(Stamp)
	Date = time.Unix(int64(stamp), 0)

	date := Date.UTC().Format("2006-01-02 15:04:05 MST")
	LongVersion = fmt.Sprintf(`syncthing %s (%s %s-%s) %s@%s %s`, Version, runtime.Version(), runtime.GOOS, runtime.GOARCH, User, Host, date)

	if len(Tags) > 0 {
		LongVersion = fmt.Sprintf("%s [%s]", LongVersion, strings.Join(Tags, ", "))
	}
}