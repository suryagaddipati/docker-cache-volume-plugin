package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type rootDirs struct {
	lower  string
	upper  string
	work   string
	merged string
}

func newRootDirs(lower, upper, work, merged string) rootDirs {
	return rootDirs{
		lower:  lower,
		upper:  upper,
		work:   work,
		merged: merged,
	}
}
func (rootDirs rootDirs) mkdirs() error {
	return mkdirs(rootDirs.lower, rootDirs.upper, rootDirs.work, rootDirs.merged)
}

type cacheDriver struct {
	sync.RWMutex
	rootDirs *rootDirs
}

func newCacheDriver(lower, upper, work, merged string) (*cacheDriver, error) {
	rootDirs := newRootDirs(lower, upper, work, merged)
	driver := &cacheDriver{
		rootDirs: &rootDirs,
	}
	err := rootDirs.mkdirs()
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func (driver *cacheDriver) Create(req *volume.CreateRequest) error {
	logrus.WithField("method", "create").Debugf("%#v", req)

	driver.Lock()
	defer driver.Unlock()
	jobName, buildNumber, err := getNames(req.Name)
	if err != nil {
		return err
	}
	buildVolume := newBuildVolume(jobName, buildNumber, driver.rootDirs)
	if buildVolume.exists() {
		return logError("Create-%s: The volume already exists", req.Name)
	}
	if err := buildVolume.init(); err != nil {
		return logError("Create-%s: Failed to create Dirs. %s", req.Name, err)
	}
	return nil
}

func (driver *cacheDriver) Remove(req *volume.RemoveRequest) error {
	logrus.WithField("method", "remove").Debugf("%#v", req)

	driver.Lock()
	defer driver.Unlock()

	jobName, buildNumber, _ := getNames(req.Name)
	buildVolume := newBuildVolume(jobName, buildNumber, driver.rootDirs)
	return buildVolume.cleanUpVolume()
}

func (driver *cacheDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	logrus.WithField("method", "path").Debugf("%#v", req)

	driver.RLock()
	defer driver.RUnlock()
	jobName, buildNumber, err := getNames(req.Name)
	if err != nil {
		return &volume.PathResponse{}, err
	}

	return &volume.PathResponse{Mountpoint: path.Join(driver.rootDirs.merged, jobName, buildNumber)}, nil

}

func (driver *cacheDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	logrus.WithField("method", "mount").Debugf("%#v", req)

	driver.Lock()
	defer driver.Unlock()

	jobName, buildNumber, err := getNames(req.Name)
	if err != nil {
		return &volume.MountResponse{}, err
	}
	if err := newBuildVolume(jobName, buildNumber, driver.rootDirs).mount(); err != nil {
		return &volume.MountResponse{}, err
	}

	return &volume.MountResponse{Mountpoint: path.Join(driver.rootDirs.merged, jobName, buildNumber)}, nil
}

func (driver *cacheDriver) Unmount(req *volume.UnmountRequest) error {
	logrus.WithField("method", "unmount").Debugf("%#v", req)

	driver.Lock()
	defer driver.Unlock()

	jobName, buildNumber, err := getNames(req.Name)
	if err != nil {
		return err
	}
	buildVolume := newBuildVolume(jobName, buildNumber, driver.rootDirs)

	if err := buildVolume.destroy(); err != nil {
		return err
	}

	return buildVolume.cleanUpVolume()
}

func (d *cacheDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	logrus.WithField("method", "get").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	jobName, buildNumber, err := getNames(r.Name)
	if err != nil {
		return nil, err
	}
	buildVolume := newBuildVolume(jobName, buildNumber, d.rootDirs)

	if buildVolume.exists() {
		return &volume.GetResponse{Volume: d.volume(jobName, buildNumber)}, nil

	}
	return &volume.GetResponse{}, logError("volume %s not found", r.Name)
}

func (driver *cacheDriver) List() (*volume.ListResponse, error) {
	logrus.WithField("method", "list").Debugf("")

	driver.Lock()
	defer driver.Unlock()

	merged := driver.rootDirs.merged
	matches, err := filepath.Glob(fmt.Sprintf("%s/*/*", merged))
	if err != nil {
		return &volume.ListResponse{}, err
	}
	if matches != nil {
		var volumes []*volume.Volume = make([]*volume.Volume, len(matches))
		for i, match := range matches {
			mergeDir := strings.Replace(match, merged+"/", "", -1)
			dirs := strings.Split(mergeDir, "/")
			volumes[i] = driver.volume(dirs[0], dirs[1])
		}
		return &volume.ListResponse{Volumes: volumes}, nil

	}
	return &volume.ListResponse{}, nil
}

func (d *cacheDriver) Capabilities() *volume.CapabilitiesResponse {
	logrus.WithField("method", "capabilities").Debugf("")
	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "local"}}
}

func (driver *cacheDriver) volume(jobName, buildNumber string) *volume.Volume {
	return &volume.Volume{
		Name:       jobName + "-" + buildNumber,
		Mountpoint: path.Join(driver.rootDirs.merged, jobName, buildNumber),
	}
}

func getNames(volumeName string) (string, string, error) {
	names := strings.Split(volumeName, "-")
	if len(names) > 1 {
		return names[0], names[1], nil
	}
	return "", "", errors.New(volumeName + " is not valid.")
}

func logError(format string, args ...interface{}) error {
	logrus.Errorf(format, args...)
	return fmt.Errorf(format, args)
}
