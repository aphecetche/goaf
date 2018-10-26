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

var directory string
var pattern string
var prefix string

// Report duplicate files
func reportDuplicates(fis fstat.FileInfoSlice) {
	defer TimeTrack(time.Now(), "reportDuplicates")
	encountered := map[string]int{}
	for i := 0; i < len(fis); i++ {
		encountered[fis[i].Path()]++
	}
	for h := range encountered {
		if encountered[h] > 1 {
			fmt.Printf("%s appears %d times\n", h, encountered[h]-1)
		}
	}
	fmt.Printf("%d file scanned\n", len(encountered))
}

func reportGroups(fis fstat.FileInfoSlice) {
	defer TimeTrack(time.Now(), "reportGroups")

	simperiods := fstat.NewBag("SIM-PERIOD")
	dataperiods := fstat.NewBag("DATA-PERIOD")
	datatype := fstat.NewBag("DATA-TYPE")
	hosts := fstat.NewBag("HOST")
	passes := fstat.NewBag("PASS")
	bags := []*fstat.Bag{simperiods, dataperiods, hosts, datatype, passes}

	for i := 0; i < len(fis); i++ {

		period := fis[i].Period()

		if len(period) == 0 && !fis[i].IsUser() {
			fstat.Dump(fis[i])
			continue
		}

		if fis[i].IsSim() {
			simperiods.Add(period, &fis[i])
		}
		if fis[i].IsData() {
			dataperiods.Add(period, &fis[i])
		}

		dt := fis[i].DataType()

		datatype.Add(dt, &fis[i])

		if dt == "ESD" || dt == "SIM-" || dt == "DATA-" {
			fstat.Dump(fis[i])
			os.Exit(42)
		}

		if len(fis[i].Pass()) > 0 {
			pass := period
			pass += "/"
			pass += fis[i].Pass()
			passes.Add(pass, &fis[i])
		}

		hosts.Add(fis[i].Host(), &fis[i])
	}

	hostnames := hosts.HostNames()

	for i := 0; i < len(bags); i++ {
		if bags[i].Tag() == "PASS" {
			bags[i].Show(hostnames)
		} else {
			bags[i].Show([]string{})
		}
	}
}

// Report generates HTML pages
func report(cmd *cobra.Command, args []string) {

	files := getFileNames()
	fileinfos := processFiles(files)

	reportGroups(fileinfos)

	all := fstat.NewFileInfoGroup(fileinfos, "all")
	size := all.Size()
	fmt.Printf("%v\n", all)
	fmt.Printf("Total size %d (%d GB)\n", size, size/1024/1024/1024)

	reportDuplicates(fileinfos)
}

func parseInfo(line string) *fstat.FileInfo {
	s := strings.SplitN(line, " ", 5)
	lastmod, _ := strconv.ParseInt(s[0], 10, 64)
	lastacc, _ := strconv.ParseInt(s[1], 10, 64)
	size, _ := strconv.ParseInt(s[2], 10, 64)
	path := s[3][len(prefix):]
	host := s[4][4:] //FIXME: assuming here the SAF- prefix is put in front of the hostname !
	return fstat.NewFileInfo(size, path, host, lastmod, lastacc)
}

func linesToFileInfos(lines []string) fstat.FileInfoSlice {
	defer TimeTrack(time.Now(), "linesToFileInfos")

	fileinfos := make(fstat.FileInfoSlice, len(lines))
	for i := 0; i < len(lines); i++ {
		fileinfos[i] = *parseInfo(lines[i])
	}
	return fileinfos
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func getFileNames() []string {
	dir := fmt.Sprintf("%s/%s", directory, pattern)
	files, err := filepath.Glob(dir)
	if err != nil {
		log.Fatal(err.Error())
	}
	return files
}

// Process the file list, using one go routine per file
func processFiles(files []string) []fstat.FileInfo {
	var wg sync.WaitGroup

	fileinfos := make([][]fstat.FileInfo, len(files))

	for i, file := range files {
		wg.Add(1)
		go func(i int, file string) {
			defer wg.Done()
			fileinfos[i] = getFileInfos(file)
		}(i, file)
	}
	wg.Wait()

	n := 0
	for i, _ := range files {
		n += len(fileinfos[i])
	}

	fis := make([]fstat.FileInfo, n)

	n = 0
	for i, _ := range files {
		for j := 0; j < len(fileinfos[i]); j++ {
			fis[n] = fileinfos[i][j]
			n++
		}
	}
	return fis
}

func getFileInfos(file string) []fstat.FileInfo {
	filelines, _ := readfile(file)
	return linesToFileInfos(filelines)
}

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
	return lines[:len(lines)-1], nil
}
