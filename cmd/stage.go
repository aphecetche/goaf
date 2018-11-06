package cmd

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aphecetche/goaf/fstat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	bolt "go.etcd.io/bbolt"
)

var stageCmd = &cobra.Command{
	Use:   "stage",
	Short: "Handle staging",
	Run:   stage,
}

func addToStore(db *bolt.DB, servers []string, files map[string][]string) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	for _, server := range servers {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(server))
			for _, file := range files[server] {
				fi := fstat.NewFileInfoBare(file, server)
				buf.Reset()
				err := enc.Encode(fi)
				if err != nil {
					fmt.Errorf("encode error:%s b:%s", err, buf)
				}
				err = b.Put([]byte(file), buf.Bytes())
				if err != nil {
					return fmt.Errorf("insert file: %s", err)
				}
			}
			return nil
		})
	}
}

func getDB(servers []string) *bolt.DB {
	db, err := bolt.Open("requests.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		for _, server := range servers {
			_, err := tx.CreateBucket([]byte(server))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
		}
		return nil

	})

	return db
}

func showRequest() {

	for _, server := range servers {
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(server))

			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				buf := bytes.NewReader(v)
				dec := gob.NewDecoder(buf)
				var fi fstat.FileInfo
				err := dec.Decode(&fi)
				if err != nil {
					// fmt.Printf("key=%s err=%s", k, err)
					fmt.Errorf("decode error:%s b:%s", err, buf)
				}
				fmt.Printf("key=%s, value=%s\n", k, fi.String())
			}

			return nil
		})
	}
}

func addRequest() {

	files := readRequest(addRequestFlag, servers)

	addToStore(db, servers, files)

}

func buildFileName(file, filter, aliphysics string) string {
	if len(filter) == 0 {
		return file
	}
	return fmt.Sprintf("%s.FILTER_%s_WITH_ALIPHYSICS_%s.root", strings.TrimSuffix(file, ".root"), filter, aliphysics)
}

func readRequest(filelist string, servers []string) map[string][]string {

	file, err := os.Open(filelist)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	n := 0

	files := make(map[string][]string, len(servers))

	for _, s := range servers {
		files[s] = []string{}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key := strings.Trim(scanner.Text(), " ")
		if strings.HasPrefix(key, "/alice") {
			files[servers[n]] = append(files[servers[n]], buildFileName(key, filterFlag, aliphysicsFlag))
			n++
			if n == len(servers) {
				n = 0
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return files
}

func stage(cmd *cobra.Command, args []string) {
	servers = viper.GetStringSlice("servers")
	db = getDB(servers)
	defer db.Close()

	if len(addRequestFlag) > 0 {
		addRequest()
	}
	if showRequestFlag {
		showRequest()
	}
}

var addRequestFlag string
var showRequestFlag bool
var filterFlag string
var aliphysicsFlag string
var servers []string
var db *bolt.DB

func init() {
	RootCmd.AddCommand(stageCmd)
	stageCmd.PersistentFlags().StringVarP(&addRequestFlag, "add-request", "r", "", "filelist of the files to be requestd")
	stageCmd.PersistentFlags().StringVarP(&filterFlag, "with-filter", "f", "", "filter to be used")
	stageCmd.PersistentFlags().StringVarP(&aliphysicsFlag, "with-aliphysics", "a", "v5-09-39-01-1", "AliPhysics version to be used")
	stageCmd.PersistentFlags().BoolVarP(&showRequestFlag, "show-request", "s", false, "show pending staging requests")
}
