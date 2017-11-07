package main

import (
  "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
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


func (d *cacheDriver) Create(r *volume.CreateRequest) error {
	logrus.WithField("method", "create").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()
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

  return nil, nil
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
