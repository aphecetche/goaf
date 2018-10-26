package fstat

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gonum/stat"
)

type Bag struct {
	m   map[string]*FileInfoGroup
	v   []*FileInfoGroup
	tag string
}

func (b *Bag) HostNames() []string {
	hostnames := make([]string, len(b.m))
	i := 0
	for h := range b.m {
		hostnames[i] = h
		i++
	}
	return hostnames
}

func (b *Bag) Tag() string {
	return b.tag
}

func NewBag(tag string) *Bag {
	m := make(map[string]*FileInfoGroup)
	v := []*FileInfoGroup{}
	return &Bag{m, v, tag}
}

func (b *Bag) Add(label string, fi *FileInfo) {
	fig, ok := b.m[label]
	if !ok {
		fig = NewFileInfoGroup(FileInfoSlice{}, label)
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

func (b Bag) Print(splitByHost []string) {
	for _, fig := range b.v {
		fmt.Printf("%v", fig)
		if len(splitByHost) > 0 {
			byhost := fig.SplitByHost(splitByHost)

			sizes := make([]float64, len(byhost))
			for i, bh := range byhost {
				sizes[i] = float64(bh.Size())
			}

			mean, disp := stat.MeanStdDev(sizes, nil)

			x := 100.0 * disp / mean
			fmt.Printf(" RDisp %3.0f %%", x)

			fmt.Printf(" GB/host ")
			for _, bh := range byhost {
				fmt.Printf("%10.2f", bh.SizeInGB())
			}
		}
		fmt.Println()
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

func (b Bag) Show(splitByHost []string) {
	name := b.tag
	if len(b.v) > 1 && name[len(name)-1] != 'S' {
		name += "s"
	}
	sort.Strings(splitByHost)
	fmt.Printf("--- %d %s - %d files - %7.2f GB\n", len(b.v), strings.ToUpper(name),
		b.NumberOfFiles(), b.SizeInGB())
	if len(splitByHost) > 0 {
		fmt.Printf(strings.Repeat(" ", 89))
		for _, h := range splitByHost {
			fmt.Printf("%10s", h)
		}
		fmt.Println()
	}
	b.SortBySize()
	b.Print(splitByHost)
}
