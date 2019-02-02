# Creaper (Cache Reaper)

![Screenshot](https://media.giphy.com/media/aReX2IJ3rtVeg/giphy.gif)

Utility application that monitors a specified cache directory and deletes least-recently-accessed 
files if the cache directory gets too large.

```
Usage of ./creaper:
  -cacheDir string
    	(required) cache directory to monitor
  -frequency duration
    	frequency with which the creaper should check disk usage (default 1m0s)
  -maxSize string
    	max cache size (examples: 1048576k, 1024m, 1g) (default "1g")
  -force
    	turn off sanity checking
```
