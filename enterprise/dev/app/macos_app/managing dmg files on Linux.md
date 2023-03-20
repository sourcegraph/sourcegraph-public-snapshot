# managing dmg files on Linux

## Ubuntu VM

```bash
apt update && apt upgrade && apt install hfsprogs kpartx


```

We don't really want to build the dmg on Linux, because it won't have the dimensions and icon size + placement.
The background we can manage by placing the file in the .background folder and copying the .DS_Store file to point to it,
but I don't know how to duplicate the size and the icon placement.
We use an applescript to specify the dimensions and icon size + placement.

Creating the dmg using `hdiutils` also creates a more "sophisticated" image, with two partitions. Creating the dmg using `mkfs.hfsplus` from `hfsprogs` on Linux creates one with just one partition. It seems to work ok so far, but I haven't tried compressing it yet.

`mkfs.hfsplus` works in Docker containers, but I can't get it to `mount` the image.
