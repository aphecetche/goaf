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

package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/aphecetche/goaf/cmd"
)

func main() {
	fcpu, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(fcpu); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	defer cmd.TimeTrack(time.Now(), "main")

	cmd.Execute()

	fmem, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	//	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(fmem); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	fmem.Close()
}
