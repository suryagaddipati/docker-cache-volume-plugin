package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"log"
	"os"
	"strconv"
)

const socketAddress = "/run/docker/plugins/cache-driver.sock"

func main() {
	debug := os.Getenv("DEBUG")
	if ok, _ := strconv.ParseBool(debug); ok {
		logrus.SetLevel(logrus.DebugLevel)
	}
	lowerRootDir := "/cache"
	upperRootDir := "/mnt/cache-upper"
	workRootDir := "/mnt/cache-work"
	mergedRootDir := "/mnt/cache-merged"

	d, err := newCacheDriver(lowerRootDir, upperRootDir, workRootDir, mergedRootDir)
	if err != nil {
		log.Fatal(err)
	}
	h := volume.NewHandler(d)
	logrus.Infof("listening on %s", socketAddress)
	logrus.Error(h.ServeUnix(socketAddress, 0))
}
