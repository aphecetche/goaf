package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

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
	for _, server := range servers {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(server))
			for _, file := range files[server] {
				err := b.Put([]byte(file), []byte(filterFlag))
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

func request() {

	servers := viper.GetStringSlice("servers")

	db := getDB(servers)
	defer db.Close()

	files := readRequest(requestFlag, servers)

	addToStore(db, servers, files)

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
			files[servers[n]] = append(files[servers[n]], key)
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
	if len(requestFlag) > 0 {
		request()
	}
}

var requestFlag string
var filterFlag string
var aliphysicsFlag string

func init() {
	RootCmd.AddCommand(stageCmd)
	stageCmd.PersistentFlags().StringVarP(&requestFlag, "add-request", "r", "", "filelist of the files to be requestd")
	stageCmd.PersistentFlags().StringVarP(&filterFlag, "with-filter", "f", "NONE", "filter to be used")
	stageCmd.PersistentFlags().StringVarP(&aliphysicsFlag, "with-aliphysics", "a", "v5-09-39-01-1", "AliPhysics version to be used")
}
