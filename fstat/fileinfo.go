package fstat

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
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

	parts     []string
	israw     bool
	pass      string
	runNumber int
	period    string
}

type PrivateFileInfo struct {
	Size, LastMod, LastAcc int64
	Path, Host             string
}

func (fi FileInfo) GobEncode() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(PrivateFileInfo{Size: fi.size, LastMod: fi.lastmod, LastAcc: fi.lastacc, Path: fi.path, Host: fi.host})
	if err != nil {
		log.Fatal("encoding error:", err)
	}
	return b.Bytes(), nil
}

// UnmarshalBinary modifies the receiver so it must take a pointer receiver.
func (fi *FileInfo) GobDecode(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	var pfi PrivateFileInfo
	err := dec.Decode(&pfi)
	if err != nil {
		fi = NewFileInfo(pfi.Size, pfi.Path, pfi.Host, pfi.LastMod, pfi.LastAcc)
	}
	return err
}

func NewFileInfoBare(path, hostname string) *FileInfo {
	fi := FileInfo{path: path, host: hostname}
	fi.parts = strings.Split(path, "/")
	fi.buildPass()
	fi.buildRunNumber()
	fi.buildPeriod()
	fi.buildIsRaw()
	return &fi
}

func NewFileInfo(size int64, path, hostname string, lastmod, lastacc int64) *FileInfo {
	fi := FileInfo{size: size, path: path, host: hostname, lastmod: lastmod, lastacc: lastacc}
	fi.parts = strings.Split(path, "/")
	fi.buildPass()
	fi.buildRunNumber()
	fi.buildPeriod()
	fi.buildIsRaw()
	return &fi
}

func (fi FileInfo) Path() string {
	return fi.path
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

func IsSameFile(a, b *FileInfo) bool {
	if len(a.path) != len(b.path) {
		return false
	}
	if a.path == b.path {
		return true
	}
	return false
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
	return fi.parts[len(fi.parts)-1]
}

func (fi FileInfo) IsGroup() bool {
	return strings.HasPrefix(fi.parts[1], "PWG")
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

func (fi FileInfo) IsESD() bool {
	return strings.Contains(fi.path, "AliESD")
}

func (fi FileInfo) IsAOD() bool {
	return strings.Contains(fi.path, "AliAOD")
}

func (fi FileInfo) IsRaw() bool {
	return fi.israw
}

func (fi *FileInfo) buildIsRaw() {
	fi.israw = false
	for _, p := range fi.parts {
		if p == "raw" && !fi.IsESD() && !fi.IsAOD() {
			fi.israw = true
		}
	}
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
	return fi.pass
}

func (fi *FileInfo) buildPass() {
	fi.pass = ""
	if !fi.IsData() {
		return
	}

	for i, p := range fi.parts {
		if isPeriod(p) {
			fi.pass = fi.parts[i+2]
			if strings.HasPrefix(fi.parts[i+3], "AOD") {
				fi.pass += "/" + fi.parts[i+3]
			}
			return
		}
	}
}
func (fi FileInfo) RunNumber() int {
	return fi.runNumber
}

func (fi *FileInfo) buildRunNumber() {
	fi.runNumber = -1
	for _, p := range fi.parts {
		if len(p) != 9 && len(p) != 6 {
			continue
		}
		if n, err := strconv.Atoi(p); err == nil {
			fi.runNumber = n
		}
	}
}

func (fi FileInfo) Period() string {
	return fi.period
}

func (fi *FileInfo) buildPeriod() {
	fi.period = ""
	for _, p := range fi.parts {
		if isPeriod(p) {
			fi.period = p
			return
		}
	}
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

func (fig *FileInfoGroup) AppendFileInfo(fi *FileInfo) {
	fig.FileInfoSlice = append(fig.FileInfoSlice, *fi)
	fig.size += fi.Size()
}

func (fig FileInfoGroup) String() string {
	return fmt.Sprintf("%30s : %7.2f GB (%6d files) %4.0f days", fig.Label(), fig.SizeInGB(), len(fig.FileInfoSlice), fig.AgeInDays())
}

// AgeInDays returns the mean number of days since the creation
// of the file group
func (fig FileInfoGroup) AgeInDays() float64 {
	var d int64 = 0
	for _, f := range fig.FileInfoSlice {
		d += f.lastmod
	}
	d /= int64(len(fig.FileInfoSlice))
	t := time.Unix(d, 0)
	return time.Since(t).Hours() / 24
}

func (fig FileInfoGroup) SplitByHost(hosts []string) []*FileInfoGroup {
	bh := make([]*FileInfoGroup, len(hosts))
	for i, h := range hosts {
		fi := FileInfoSlice{}
		for _, f := range fig.FileInfoSlice {
			if f.Host() == h {
				fi = append(fi, f)
			}
		}
		bh[i] = NewFileInfoGroup(fi, h)
	}
	return bh
}
