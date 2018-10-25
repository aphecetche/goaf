// Copyright © 2016 Laurent Aphecetche
// {{ .copyright }}
//
//  This file is part of {{ .appName }}.
//
//  {{ .appName }} is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Lesser General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  {{ .appName }} is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Lesser General Public License for more details.
//
//  You should have received a copy of the GNU Lesser General Public License
//  along with {{ .appName }}. If not, see <http://www.gnu.org/licenses/>.
//

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aphecetche/goaf/fstat"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate HTML reports",
	Long:  `Generate HTML reports of various forms, e.g. pies, treemap, etc...`,
	Run:   report,
}

type Bag struct {
	m   map[string]*fstat.FileInfoGroup
	v   []*fstat.FileInfoGroup
	tag string
}

func NewBag(tag string) *Bag {
	m := make(map[string]*fstat.FileInfoGroup)
	v := []*fstat.FileInfoGroup{}
	return &Bag{m, v, tag}
}

func (b *Bag) Add(label string, fi fstat.FileInfo) {
	fig, ok := b.m[label]
	if !ok {
		fig = fstat.NewFileInfoGroup(fstat.FileInfoSlice{}, label)
		b.m[label] = fig
		b.v = append(b.v, fig)
	}
	fig.AppendFileInfo(fi)
}

func (b *Bag) SortBySize() {
	sort.Slice(b.v, func(i, j int) bool {
		return b.v[i].Size() < b.v[j].Size()
	})
}

func (b Bag) Print() {
	for _, fig := range b.v {
		fmt.Printf("%v\n", fig)
	}
}

func (b Bag) Size() int64 {
	var size int64 = 0
	for _, fig := range b.v {
		size += fig.Size()
	}
	return size
}

func (b Bag) SizeInGB() float32 {
	return (float32)(b.Size()) / 1024 / 1024 / 1024
}

func (b Bag) NumberOfFiles() int {
	n := 0
	for _, fig := range b.v {
		n += len(fig.FileInfoSlice)
	}
	return n
}

func (b Bag) Show() {
	name := b.tag
	if len(b.v) > 1 {
		name += "s"
	}
	fmt.Printf("--- %d %s - %d files - %7.2f GB\n", len(b.v), strings.ToUpper(name),
		b.NumberOfFiles(), b.SizeInGB())
	b.SortBySize()
	b.Print()
}

// Report generates HTML pages
func report(cmd *cobra.Command, args []string) {

	fileinfos := getInfos(getFileLines())
	all := fstat.NewFileInfoGroup(fileinfos, "all")

	simperiods := NewBag("SIM-PERIOD")
	dataperiods := NewBag("DATA-PERIOD")
	datatype := NewBag("DATA-TYPE")
	hosts := NewBag("HOST")
	passes := NewBag("PASS")

	bags := []*Bag{simperiods, dataperiods, hosts, datatype, passes}

	for _, f := range fileinfos {

		period := f.Period()

		if len(period) == 0 && !f.IsUser() {
			fstat.Dump(f)
			continue
		}

		if f.IsSim() {
			simperiods.Add(period, f)
		}
		if f.IsData() {
			dataperiods.Add(period, f)
		}
		datatype.Add(f.DataType(), f)

		if f.DataType() == "ESD" || f.DataType() == "SIM-" || f.DataType() == "DATA-" {
			fstat.Dump(f)
			os.Exit(42)
		}

		pass := period
		period += "/"
		period += pass

		passes.Add(pass, f)

		hosts.Add(f.Host(), f)

	}

	size := all.Size()

	fmt.Printf("%v\n", all)
	fmt.Printf("Total size %d (%d GB)\n", size, size/1024/1024/1024)

	for _, b := range bags {
		b.Show()
	}

}

func parseInfo(line string) *fstat.FileInfo {
	s := strings.Split(line, " ")
	lastmod, _ := strconv.ParseInt(s[0], 10, 64)
	lastacc, _ := strconv.ParseInt(s[1], 10, 64)
	size, _ := strconv.ParseInt(s[2], 10, 64)
	path := strings.Replace(s[3], prefix, "", 1)
	host := s[4]
	return fstat.NewFileInfo(size, path, host, lastmod, lastacc)
}

func getInfos(lines []string) fstat.FileInfoSlice {
	defer timeTrack(time.Now(), "getInfos")

	fileinfos := make(fstat.FileInfoSlice, len(lines))
	for i, l := range lines {
		fileinfos[i] = *parseInfo(l)
	}
	return fileinfos
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

// Get the file list, using one go routine per file, no channels
func getFileLines() []string {

	defer timeTrack(time.Now(), "getfilelist1")

	dir := fmt.Sprintf("%s/%s", directory, pattern)

	files, err := filepath.Glob(dir)

	if err != nil {
		log.Fatal(err.Error())
	}

	var filelines [][]string
	var lines []string

	filelines = make([][]string, len(files))

	var wg sync.WaitGroup

	for i, file := range files {
		wg.Add(1)
		go func(i int, file string) {
			defer wg.Done()
			filelines[i], _ = readfile(file)
		}(i, file)
	}
	wg.Wait()

	lines = []string{}
	for i := range files {
		lines = append(lines, filelines[i]...)
	}

	log.Printf("getfilelist1 : number of lines : %d\n", len(lines))

	return lines
}

var directory string
var pattern string
var prefix string
var wg3 sync.WaitGroup

// The init function defines the flags we use for this command
func init() {
	RootCmd.AddCommand(reportCmd)
	reportCmd.PersistentFlags().StringVarP(&directory, "directory", "d", "", "source directory where to find the file lists")
	reportCmd.PersistentFlags().StringVarP(&pattern, "filepattern", "f", "", "starting part of the filenames to look for into directory")
	reportCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "/data", "prefix to strip from the fullpath of the results of the find command")
}

// Read a file and store it into a string slice
func readfile(filename string) ([]string, error) {

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	fmt.Printf("readfile %s nlines %d\n", filename, len(lines))
	return lines[:len(lines)-1], nil
}
