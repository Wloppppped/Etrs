package etrs_ser

import (
	"log"
	"os"

	etrs "Etrs/internal"
)

func ClientInit() *etrs.EtrsResolver {
	// etcd, username, passwd := readEnv()

	// er, err := etrs.EtrsResolverInit([]string{etcd}, etrs.WithAuthOpt(username, passwd))
	er, err := etrs.EtrsResolverInit([]string{"127.0.0.1:2379"}, etrs.WithAuthOpt("root", "onxM8vRpTD"))
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
