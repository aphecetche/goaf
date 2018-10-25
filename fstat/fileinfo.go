package fstat

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FileInfo is a struct to hold basic information about a file : size, path,
// timestamps and origin (i.e. hostname)
type FileInfo struct {
	size    int64
	path    string
	host    string
	lastmod int64
	lastacc int64
}

func NewFileInfo(size int64, path, hostname string, lastmod, lastacc int64) *FileInfo {
	return &FileInfo{size, path, hostname, lastmod, lastacc}
}

func isPeriod(s string) bool {
	return strings.HasPrefix(s, "LHC")
}

func intToDate(d int64) string {
	t := time.Unix(d, 0)
	return t.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
}

func (fi FileInfo) Size() int64 {
	return fi.size
}

func (fi FileInfo) Host() string {
	return fi.host
}

func Dump(f FileInfo) {
	fmt.Printf("\n%v\n", f)
	fmt.Printf("run=%09d data=%v sim=%v period=%v pass=%v isuser=%v user=%s israw=%v isesd=%v isaod=%v isgroup=%v\n", f.RunNumber(), f.IsData(), f.IsSim(), f.Period(), f.Pass(), f.IsUser(), f.UserName(), f.IsRaw(),
		f.IsESD(), f.IsAOD(), f.IsGroup())
}
func (fi FileInfo) String() string {
	return fmt.Sprintf("{path:%s host:[%s] size:(%d) mod:%s acc:%s}", fi.path, fi.host, fi.size, intToDate(fi.lastmod), intToDate(fi.lastacc))
}

func (fi FileInfo) DataType() string {
	var dt string
	if fi.IsData() {
		dt += "DATA-"
	}
	if fi.IsSim() {
		dt += "SIM-"
	}
	if fi.IsUser() {
		dt += "USER-"
	}
	if fi.IsGroup() {
		dt += "GROUP-"
	}
	if fi.IsESD() {
		dt += "ESD"
	}
	if fi.IsAOD() {
		dt += "AOD"
	}
	if fi.IsRaw() {
		dt += "RAW"
	}
	if fi.IsQA() {
		dt += "QA"
	}
	if fi.IsGene() {
		dt += "GENE"
	}
	return dt
}

func (fi FileInfo) FileName() string {
	sp := fi.SplitPath()
	return sp[len(sp)-1]
}

func (fi FileInfo) IsGroup() bool {
	sp := fi.SplitPath()
	return strings.HasPrefix(sp[1], "PWG")
}

func (fi FileInfo) IsGene() bool {
	return strings.Contains(fi.FileName(), "pyxsec")
}

func (fi FileInfo) IsQA() bool {
	return strings.Contains(fi.FileName(), "QA")
}

func (fi FileInfo) IsFiltered() bool {
	return strings.Contains(fi.path, "FILTER")
}

func (fi FileInfo) SplitPath() []string {
	return strings.Split(fi.path, "/")
}

func (fi FileInfo) IsESD() bool {
	return strings.Contains(fi.path, "AliESD")
}

func (fi FileInfo) IsAOD() bool {
	return strings.Contains(fi.path, "AliAOD")
}

func (fi FileInfo) IsRaw() bool {
	for _, p := range fi.SplitPath() {
		if p == "raw" && !fi.IsESD() && !fi.IsAOD() {
			return true
		}
	}
	return false
}

func (fi FileInfo) IsUser() bool {
	return strings.HasPrefix(fi.path, "/alice/cern.ch/user")
}

func (fi FileInfo) UserName() string {
	if !fi.IsUser() {
		return ""
	}
	u := strings.Replace(fi.path, "/alice/cern.ch/user/", "", 1)
	return strings.Split(u, "/")[1]
}

func (fi FileInfo) IsData() bool {
	return strings.HasPrefix(fi.path, "/alice/data")
}

func (fi FileInfo) IsSim() bool {
	return strings.HasPrefix(fi.path, "/alice/sim")
}

func (fi FileInfo) Pass() string {
	if !fi.IsData() {
		return ""
	}

	parts := fi.SplitPath()
	for i, p := range parts {
		if isPeriod(p) {
			pass := parts[i+2]
			if strings.HasPrefix(parts[i+3], "AOD") {
				pass += "/" + parts[i+3]
			}
			return pass
		}
	}
	return ""
}

func (fi FileInfo) RunNumber() int {
	for _, p := range fi.SplitPath() {
		if len(p) != 9 && len(p) != 6 {
			continue
		}
		if n, err := strconv.Atoi(p); err == nil {
			return n
		}
	}
	return -1
}

func (fi FileInfo) Period() string {
	for _, p := range fi.SplitPath() {
		if isPeriod(p) {
			return p
		}
	}
	return ""
}

type FileInfoSlice []FileInfo

type FileInfoGroup struct {
	FileInfoSlice
	label string
	size  int64
}

func NewFileInfoGroup(fis FileInfoSlice, label string) *FileInfoGroup {
	var size int64 = 0
	for _, f := range fis {
		size += f.Size()
	}
	return &FileInfoGroup{fis, label, size}
}

func (fig FileInfoGroup) Size() int64 {
	return fig.size
}

func (fig FileInfoGroup) SizeInGB() float32 {
	return float32(fig.size) / 1024 / 1024 / 1024
}

func (fig FileInfoGroup) Label() string {
	return fig.label
}

func (fig *FileInfoGroup) AppendFileInfo(fi FileInfo) {
	fig.FileInfoSlice = append(fig.FileInfoSlice, fi)
	fig.size += fi.Size()
}

func (fig FileInfoGroup) String() string {
	return fmt.Sprintf("%30s : %7.2f GB (%6d files)", fig.Label(), fig.SizeInGB(), len(fig.FileInfoSlice))
}
