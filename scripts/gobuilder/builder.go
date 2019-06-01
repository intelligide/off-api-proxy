package gobuilder

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	Debug = os.Getenv("BUILDDEBUG") != ""
)

type Builder struct {
	targets map[string]Target
	GoOS string
	Cc string
	GoArch string
	GoCmd string
	Race bool
	PkgDir string
	InstallSuffix string
	DebugBinary bool
	Tags []string
	LdFlags []string
	Coverage bool
	NoUpgrade bool
	Timeout uint
	Jobs uint
}

func NewBuilder() *Builder {
	return &Builder {
		targets: make(map[string]Target),
		Timeout: 120,
		Jobs: uint(runtime.NumCPU()),
	}
}

func (this *Builder) SetupFlags() {
	flag.StringVar(&this.GoArch, "goarch", runtime.GOARCH, "GOARCH")
	flag.StringVar(&this.GoOS, "goos", runtime.GOOS, "GOOS")
	flag.StringVar(&this.GoCmd, "gocmd", "go", "Specify `go` commands")
	flag.BoolVar(&this.NoUpgrade, "no-upgrade", this.NoUpgrade, "Disable upgrade functionality")
	flag.BoolVar(&this.Race, "race", this.Race, "Use race detector")
	flag.StringVar(&this.InstallSuffix, "installsuffix", this.InstallSuffix, "Install suffix, optional")
	flag.StringVar(&this.PkgDir, "pkgdir", "", "Set -pkgdir parameter for `go build_info`")
	flag.StringVar(&this.Cc, "cc", os.Getenv("CC"), "Set CC environment variable for `go build_info`")
	flag.BoolVar(&this.DebugBinary, "debug-binary", this.DebugBinary, "Create unoptimized binary to use with delve, set -gcflags='-N -l' and omit -ldflags")
	flag.BoolVar(&this.Coverage, "coverage", this.Coverage, "Write coverage profile of tests to coverage.txt")
}

func (this *Builder) AddTarget(name string, target Target) {
	name = strings.Trim(name, " ")
	if name == "all" {
		log.Fatalln(name + ": Reserved target name.")
	}
	if len(target.Name) == 0 {
		target.Name = name
	}
	if len(target.BinaryName) == 0 {
		target.BinaryName = ToKebab(target.Name)
	}
	target.Validate()
	this.targets[name] = target;
}

func (this *Builder) AddTargets(targets map[string]Target) {
	for name, target := range targets {
		this.AddTarget(name, target)
	}
}

func (this *Builder) getTarget(targetName string) Target {
	target, ok := this.targets[targetName]
	if !ok {
		log.Fatalln("Unknown target", targetName)
	}
	return target
}

/**
 * Building & Install
 */

func (this *Builder) Install(targetName string, tags []string) {
	if len(targetName) == 0 || targetName == "all" {
		if Debug {
			log.Println("Installing all targets...")
		}
		allTargetNames := make([]string, 0, len(this.targets))
		for name := range this.targets {
			allTargetNames = append(allTargetNames, name)
		}
		this.InstallMultiple(allTargetNames, tags)
		return
	}

	target := this.getTarget(targetName)
	tags = append(target.Tags, tags...)
	tags = append(tags, this.Tags...)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("GOBIN", filepath.Join(cwd, "bin"))

	args := []string{"install", "-v"}
	args = this.appendParameters(args, tags, target)

	os.Setenv("GOOS", this.GoOS)
	os.Setenv("GOARCH", this.GoArch)
	os.Setenv("CC", this.Cc)

	runPrint(this.GoCmd, args...)
}

func (this *Builder) InstallMultiple(targetNames []string, tags []string) {

	var wg sync.WaitGroup

	wg.Add(len(targetNames))

	for _, name := range targetNames {
		go func() {
			defer wg.Done()
			this.Install(name, tags)
		}()
	}
	if Debug {
		log.Println("Waiting...")
	}
	wg.Wait()
}

func (this *Builder) Build(targetName string, tags []string) {
	if len(targetName) == 0 || targetName == "all" {
		if Debug {
			log.Println("Building all targets...")
		}
		allTargetNames := make([]string, 0, len(this.targets))
		for name := range this.targets {
			allTargetNames = append(allTargetNames, name)
		}
		this.BuildMultiple(allTargetNames, tags)
		return
	}

	target := this.getTarget(targetName)
	tags = append(target.Tags, tags...)
	tags = append(tags, this.Tags...)

	rmr(this.GetTargetFullBinaryName(target))

	args := []string{"build_info", "-v"}
	if len(target.BinaryName) > 0 {
		args = append(args, "-o", "build_info/bin/" + this.GetTargetFullBinaryName(target));
	}
	args = this.appendParameters(args, tags, target)

	os.Setenv("GOOS", this.GoOS)
	os.Setenv("GOARCH", this.GoArch)
	os.Setenv("CC", this.Cc)

	runPrint(this.GoCmd, args...)
}

func (this *Builder) BuildMultiple(targetNames []string, tags []string) {
	var wg sync.WaitGroup

	wg.Add(len(targetNames))

	for _, name := range targetNames {
		go func() {
			defer wg.Done()
			this.Build(name, tags)
		}()
	}

	if Debug {
		log.Println("Waiting...")
	}
	wg.Wait()
}

func rmr(paths ...string) {
	for _, path := range paths {
		if Debug {
			log.Println("rm -r", path)
		}
		os.RemoveAll(path)
	}
}

func (this *Builder) appendParameters(args []string, tags []string, target Target) []string {
	if this.PkgDir != "" {
		args = append(args, "-pkgdir", this.PkgDir)
	}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, " "))
	}
	if this.InstallSuffix != "" {
		args = append(args, "-installsuffix", this.InstallSuffix)
	}
	if this.Race {
		args = append(args, "-race")
	}

	if !this.DebugBinary {
		ldflags := append(target.LdFlags, this.LdFlags...)
		ldflags = append(ldflags, "-w")

		// Regular binaries get version tagged and skip some debug symbols
		args = append(args, "-ldflags", strings.Join(ldflags[:], " "))
	} else {
		// -gcflags to disable optimizations and inlining. Skip -ldflags
		// because `Could not launch program: decoding dwarf section info at
		// offset 0x0: too short` on 'dlv exec ...' see
		// https://github.com/derekparker/delve/issues/79
		args = append(args, "-gcflags", "-N -l")
	}

	return append(args, target.BuildPkg)
}

/**
 * Tests & benchmarks
 */

func (this *Builder) Test(pkgs ...string) {
	args := []string{"test", "-short", "-timeout", strconv.FormatUint(uint64(this.Timeout), 10) + "s", "-tags", "purego"}

	if runtime.GOARCH == "amd64" {
		switch runtime.GOOS {
		case "darwin", "linux", "freebsd": // , "windows": # See https://github.com/golang/go/issues/27089
			args = append(args, "-race")
		}
	}

	if this.Coverage {
		args = append(args, "-covermode", "atomic", "-coverprofile", "coverage.txt")
	}

	runPrint(this.GoCmd, append(args, pkgs...)...)
}

func (this *Builder) Bench(pkgs ...string) {
	runPrint(this.GoCmd, append([]string{"test", "-run", "NONE", "-bench", "."}, pkgs...)...)
}

/**
 * Packaging
 */

func (this *Builder) BuildTar(targetName string) {
	target := this.getTarget(targetName)
	name := target.ArchiveName(this.buildArch())
	filename := name + ".tar.gz"

	var tags []string
	if this.NoUpgrade {
		tags = []string{"noupgrade"}
		name += "-noupgrade"
	}

	this.Build(targetName, tags)

	for i := range target.ArchiveFiles {
		target.ArchiveFiles[i].Src = strings.Replace(target.ArchiveFiles[i].Src, "{{binary}}", this.GetTargetFullBinaryName(target), 1)
		target.ArchiveFiles[i].Dst = strings.Replace(target.ArchiveFiles[i].Dst, "{{binary}}", this.GetTargetFullBinaryName(target), 1)
		target.ArchiveFiles[i].Dst = name + "/" + target.ArchiveFiles[i].Dst
	}

	tarGz(filename, target.ArchiveFiles)
	fmt.Println(filename)
}

func (this *Builder) BuildZip(targetName string) {
	target := this.getTarget(targetName)
	name := target.ArchiveName(this.buildArch())
	filename := name + ".zip"

	var tags []string
	if this.NoUpgrade {
		tags = []string{"noupgrade"}
		name += "-noupgrade"
	}

	this.Build(targetName, tags)

	for i := range target.ArchiveFiles {
		target.ArchiveFiles[i].Src = strings.Replace(target.ArchiveFiles[i].Src, "{{binary}}", this.GetTargetFullBinaryName(target), 1)
		target.ArchiveFiles[i].Dst = strings.Replace(target.ArchiveFiles[i].Dst, "{{binary}}", this.GetTargetFullBinaryName(target), 1)
		target.ArchiveFiles[i].Dst = name + "/" + target.ArchiveFiles[i].Dst
	}

	zipFile(filename, target.ArchiveFiles)
	fmt.Println(filename)
}

func (this *Builder) GetTargetFullBinaryName(target Target) string {
	if this.GoOS == "windows" && len(target.BinaryName) > 0{
		return target.BinaryName + ".exe"
	}
	return target.BinaryName
}

func (this *Builder) buildArch() string {
	osname := this.GoOS
	if osname == "darwin" {
		osname = "macos"
	}
	return fmt.Sprintf("%s-%s", osname, this.GoArch)
}
