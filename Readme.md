Docker volume driver to give your container exclusive view of underlying filesystem cache using [overlayfs](https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt).Think of it as `Read committed Isolation level` for your filesystem cache.

Changes made by container are written back to cache on container exit.

*Note*: There is no conflict resolution between changes made by concurrent containers, latest container to exit overwrites all previous changes.

### Installation

`docker plugin install rapt/cachedriver`

### Usage

Volume name is expected to be in `<cache-name>-<unique-id>`

eg:

```
docker run --rm -it  --name one --volume-driver rapt/cachedriver  -v foo-one:/data busybox sh
/ # ls /data/
/ #  echo one >> /data/foo-one.txt
```

An underlying cache `foo` is created at this point but contents written to this cache by the running container is not visible to concurrently running containers.
eg:
```
docker run --rm -it --name two --volume-driver rapt/cachedriver   -v foo-two:/data busybox sh
/ #  ls /data/        #foo-one.txt is not visible There

```
concurrently running containers `one` and `two` get an isolated view of the cache.

Once container `one`  exits changes made by it are written back to the cache.

```
docker run --rm -it --name three --volume-driver  rapt/cachedriver  -v foo-three:/data busybox sh
/ #  ls /data/
/ #  foo-one.txt
```
changes made by container `one` are visible to container `three` after container `one` exits.

### Use cases
* ###### Build caches in a continuous integration system.
  Build Systems like maven, sbt cache their artifacts but multiple processes cannot use a sigle underlying maven cache due to concurrency issues and exclusive locks.  Each maven process requires an isolated view of underlying cache.

  ```
  docker run --rm -it --name one --volume-driver rapt/cachedriver  -v foo-one:/~/.m2 busybox sh
  / #  mvn install  
  ```

  ```
  docker run --rm -it --name two --volume-driver  rapt/cachedriver  -v foo-two:/~/.m2 busybox sh
  / #  mvn install   # .m2 cache is primed here by one
  ```
### How does it work?

Driver uses overlayfs to provide exclusive view of  `lower` directory.

```
 docker run --rm -it --name three --volume-driver  rapt/cachedriver  -v foo-three:/data busybox sh

 / # echo foo-one > /data/foo-one.txt
 / # exit

 ```

 Inside volume driver container

 ```
  / #  docker-runc exec -t 58211af9d85c0e6e095a822f93b6522f0a3fed776c8d9bd72f62a41ebfa7e5c2 sh

  / #  mount | grep foo
overlay on /mnt/cache/merged/foo/three type overlay (rw,relatime,lowerdir=/mnt/cache/lower/foo/0,upperdir=/mnt/cache/upper/foo/three,workdir=/mnt/cache/work/foo/three)

  / #  ls /mnt/cache/lower/foo/
   0    cache-state.json

  / #  cat /mnt/cache/lower/foo/cache-state.json
   {"state":{"latest":"0"}}
 ```

After container `one` exits

```
/ # ls /mnt/cache/lower/foo/
  0     cache-state.json  three

/ # cat /mnt/cache/lower/foo/cache-state.json
{"state":{"latest":"three"}}
```
Start another container for `foo` cache
```
docker run --rm -it --name four --volume-driver  rapt/cachedriver  -v foo-four:/data busybox sh
/ # ls /data
foo-one.txt
```

Inside volume container, container `four` uses `three` as lower

```
/ # mount | grep foo
overlay on /mnt/cache/merged/foo/four type overlay (rw,relatime,lowerdir=/mnt/cache/lower/foo/three,upperdir=/mnt/cache/upper/foo/four,workdir=/mnt/cache/work/foo/four)

```
