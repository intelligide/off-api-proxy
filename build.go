package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"github.com/intelligide/off-api-proxy/scripts/gobuilder"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var targets = map[string]gobuilder.Target{
	"off-proxy": {
		BuildPkg:    "github.com/intelligide/off-api-proxy/cmd/off-proxy",
		BinaryName:  "off-proxy", // .exe will be added automatically for Windows builds
		ArchiveFiles: []gobuilder.ArchiveFile{
			{Src: "{{binary}}", Dst: "{{binary}}", Perm: 0755},
			{Src: "README.md", Dst: "README.txt", Perm: 0644},
			{Src: "LICENSE", Dst: "LICENSE.txt", Perm: 0644},
		},
		InstallationFiles: []gobuilder.ArchiveFile{
			{Src: "{{binary}}", Dst: "deb/usr/bin/{{binary}}", Perm: 0755},
			{Src: "README.md", Dst: "deb/usr/share/doc/syncthing/README.txt", Perm: 0644},
			{Src: "LICENSE", Dst: "deb/usr/share/doc/syncthing/LICENSE.txt", Perm: 0644},
		},
	},
}

var (
	versionRe        = regexp.MustCompile(`-[0-9]{1,3}-g[0-9a-f]{5,10}`)
	version          string
	debug            = os.Getenv("BUILDDEBUG") != ""
	extraTags        string
	builder          = gobuilder.NewBuilder()
)

func init() {

	builder.Tags = append(builder.Tags, "purego")

	sep := '='
	var ldflags []string
	ldflags = append(ldflags,
		fmt.Sprintf("-X github.com/intelligide/off-api-proxy/internal/build_info.Version%c%s", sep, version),
		fmt.Sprintf("-X github.com/intelligide/off-api-proxy/internal/build_info.Stamp%c%d", sep, buildStamp()),
		fmt.Sprintf("-X github.com/intelligide/off-api-proxy/internal/build_info.User%c%s", sep, buildUser()),
		fmt.Sprintf("-X github.com/intelligide/off-api-proxy/internal/build_info.Host%c%s", sep, buildHost()),
	)
}

func main() {
	log.SetFlags(0)

	builder.SetupFlags()
	flag.StringVar(&version, "version", getVersion(), "Set compiled in version string")
	flag.StringVar(&extraTags, "tags", extraTags, "Extra tags, space separated")

	flag.Parse()


	if debug {
		t0 := time.Now()
		defer func() {
			log.Println("... build_info completed in", time.Since(t0))
		}()
	}

	builder.AddTargets(targets)

	// Invoking build_info.go with no parameters at all builds everything (incrementally),
	// which is what you want for maximum error checking during development.
	if flag.NArg() == 0 {
		runCommand("install", "all")
	} else {
		targetName := ""
		if flag.NArg() > 1 {
			targetName = flag.Arg(1)
		}

		runCommand(flag.Arg(0), targetName)
	}
}

func runCommand(cmd string, targetName string) {
	switch cmd {
	case "install":
		var tags []string
		if builder.NoUpgrade {
			tags = []string{"noupgrade"}
		}
		tags = append(tags, strings.Fields(extraTags)...)
		// builder.InstallMultiple([]string { targetName }, tags)
		builder.Install(targetName, tags)

	case "build_info":
		var tags []string
		if builder.NoUpgrade {
			tags = []string{"noupgrade"}
		}
		tags = append(tags, strings.Fields(extraTags)...)
		builder.Build(targetName, tags)

	case "test":
		builder.Test("github.com/syncthing/syncthing/lib/...", "github.com/syncthing/syncthing/cmd/...")

	case "bench":
		builder.Bench("github.com/syncthing/syncthing/lib/...", "github.com/syncthing/syncthing/cmd/...")

	case "tar":
		builder.BuildTar(targetName)

	case "zip":
		builder.BuildZip(targetName)

	case "version":
		fmt.Println(getVersion())

	default:
		log.Fatalf("Unknown commands %q", cmd)
	}
}


func runError(cmd string, args ...string) ([]byte, error) {
	if debug {
		t0 := time.Now()
		log.Println("runError:", cmd, strings.Join(args, " "))
		defer func() {
			log.Println("... in", time.Since(t0))
		}()
	}
	ecmd := exec.Command(cmd, args...)
	bs, err := ecmd.CombinedOutput()
	return bytes.TrimSpace(bs), err
}

func getVersion() string {
	// First try for a RELEASE file,
	if ver, err := getReleaseVersion(); err == nil {
		return ver
	}
	// ... then see if we have a Git tag.
	if ver, err := getGitVersion(); err == nil {
		if strings.Contains(ver, "-") {
			// The version already contains a hash and stuff. See if we can
			// find a current branch name to tack onto it as well.
			return ver + getBranchSuffix()
		}
		return ver
	}
	// This seems to be a dev build_info.
	return "unknown-dev"
}

func getReleaseVersion() (string, error) {
	fd, err := os.Open("RELEASE")
	if err != nil {
		return "", err
	}
	defer fd.Close()

	bs, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(bs)), nil
}

func getGitVersion() (string, error) {
	v, err := runError("git", "describe", "--always", "--dirty")
	if err != nil {
		return "", err
	}
	v = versionRe.ReplaceAllFunc(v, func(s []byte) []byte {
		s[0] = '+'
		return s
	})
	return string(v), nil
}

func getBranchSuffix() string {
	bs, err := runError("git", "branch", "-a", "--contains")
	if err != nil {
		return ""
	}

	branches := strings.Split(string(bs), "\n")
	if len(branches) == 0 {
		return ""
	}

	branch := ""
	for i, candidate := range branches {
		if strings.HasPrefix(candidate, "*") {
			// This is the current branch. Select it!
			branch = strings.TrimLeft(candidate, " \t*")
			break
		} else if i == 0 {
			// Otherwise the first branch in the list will do.
			branch = strings.TrimSpace(branch)
		}
	}

	if branch == "" {
		return ""
	}

	// The branch name may be on the form "remotes/origin/foo" from which we
	// just want "foo".
	parts := strings.Split(branch, "/")
	if len(parts) == 0 || len(parts[len(parts)-1]) == 0 {
		return ""
	}

	branch = parts[len(parts)-1]
	switch branch {
	case "master", "release":
		// these are not special
		return ""
	}

	validBranchRe := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !validBranchRe.MatchString(branch) {
		// There's some odd stuff in the branch name. Better skip it.
		return ""
	}

	return "-" + branch
}

func buildStamp() int64 {
	// If SOURCE_DATE_EPOCH is set, use that.
	if s, _ := strconv.ParseInt(os.Getenv("SOURCE_DATE_EPOCH"), 10, 64); s > 0 {
		return s
	}

	// Try to get the timestamp of the latest commit.
	bs, err := runError("git", "show", "-s", "--format=%ct")
	if err != nil {
		// Fall back to "now".
		return time.Now().Unix()
	}

	s, _ := strconv.ParseInt(string(bs), 10, 64)
	return s
}

func buildUser() string {
	if v := os.Getenv("BUILD_USER"); v != "" {
		return v
	}

	u, err := user.Current()
	if err != nil {
		return "unknown-user"
	}
	return strings.Replace(u.Username, " ", "-", -1)
}

func buildHost() string {
	if v := os.Getenv("BUILD_HOST"); v != "" {
		return v
	}

	h, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return h
}
