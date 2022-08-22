package etrs_ser

import (
	etrs "Etrs/internal"
	"log"
	"os"
)

func ServcerInit() *etrs.EtrsRegistry {
	// etcd, username, passwd := readEnv()

	er, err := etrs.EtrsRegistryInit([]string{"127.0.0.1:2379"}, etrs.WithAuthOpt("root", "onxM8vRpTD"))
	// er, err := etrs.EtrsRegistryInit([]string{"127.0.0.1:2379"})
	if err != nil {
		log.Fatal(err)
	}

	return er
}

func readEnv() (etcd string, username string, passwd string) {
	etcd = os.Getenv("ETCD_HOST")
	username = os.Getenv("ETCD_USERNAME")
	passwd = os.Getenv("ETCD_PASSWD")
	return
}
