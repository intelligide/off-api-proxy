package gobuilder

import (
	"fmt"
	"log"
	"os"
)

type Target struct {
	Name string
	Description string
	BuildPkg string
	BinaryName string
	ArchiveFiles []ArchiveFile
	InstallationFiles []ArchiveFile
	Tags []string
	LdFlags []string
	Version string
}

func (this *Target) Validate() {
	if len(this.BuildPkg) == 0 {
		log.Fatalln(this.Name + ": Missing build_info pkg.")
	}
}

func (this *Target) ArchiveName(arch string) string {
	return fmt.Sprintf("%s-%s-%s", this.Name, arch, this.Version)
}

type ArchiveFile struct {
	Src  string
	Dst  string
	Perm os.FileMode
}