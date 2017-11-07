package main

import (
  "github.com/Sirupsen/logrus"
  "github.com/docker/go-plugins-helpers/volume"
  "sync"
  "strings"
  "errors"
  "fmt"
	"path"

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

func newCacheDriver(lower, upper, work, merged string) (*cacheDriver,error) {
  rootDirs := newRootDirs(lower, upper, work, merged)
  driver := &cacheDriver{
    rootDirs: &rootDirs,
  }
  err := rootDirs.mkdirs()
  if err!=nil{
    return nil, err
  }
  return driver,nil
}


func (driver *cacheDriver) Create(req *volume.CreateRequest) error {
  logrus.WithField("method", "create").Debugf("%#v", req)

  driver.Lock()
  defer driver.Unlock()
  jobName, buildNumber, err := getNames(req.Name)
  if err!=nil{
    return err
  }
  buildVolume := newBuildVolume(jobName, buildNumber, driver.rootDirs)
	if buildVolume.exists() {
    return logError( "Create-%s: The volume already exists", req.Name)
  }
  if err := buildVolume.init(); err != nil {
		return logError("Create-%s: Failed to create Dirs. %s", req.Name, err)
	}
  return nil
}

func (d *cacheDriver) Remove(r *volume.RemoveRequest) error {
  logrus.WithField("method", "remove").Debugf("%#v", r)

  d.Lock()
  defer d.Unlock()
  return nil
}


func (d *cacheDriver) Path(r *volume.PathRequest) (*volume.PathResponse, error) {
  logrus.WithField("method", "path").Debugf("%#v", r)

  d.RLock()
  defer d.RUnlock()
  return nil, nil
}

func (d *cacheDriver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
  logrus.WithField("method", "mount").Debugf("%#v", r)

  d.Lock()
  defer d.Unlock()
  return nil, nil
}

func (d *cacheDriver) Unmount(r *volume.UnmountRequest) error {
  logrus.WithField("method", "unmount").Debugf("%#v", r)

  d.Lock()
  defer d.Unlock()
  return nil
}

func (d *cacheDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
  logrus.WithField("method", "get").Debugf("%#v", r)

  d.Lock()
  defer d.Unlock()

  jobName, buildNumber, err := getNames(r.Name)
  if err != nil{
    return nil, err
  }
  buildVolume := newBuildVolume(jobName, buildNumber, d.rootDirs)

  if buildVolume.exists() {
    return &volume.GetResponse{ Volume: d.volume(jobName, buildNumber) }, nil

  }
  return &volume.GetResponse{}, logError("volume %s not found", r.Name)
}

func (d *cacheDriver) List() (*volume.ListResponse, error) {
  logrus.WithField("method", "list").Debugf("")

  d.Lock()
  defer d.Unlock()

  return nil, nil
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
