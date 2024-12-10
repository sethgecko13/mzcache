# mzcache

This is a caching library created as a response to this question on reddit:

https://www.reddit.com/r/golang/comments/1h4ubw9/should_i_roll_my_own_caching_server/

I had written this code a couple of years ago (not as a library) because I thought it would be simpler to write my own filesystem cache that did exactly what I wanted than it would be to figure out how to write a config file to an existing cache program or library.

Overall, I think I was correct to write it myself.

## Features

This caching library does exactly what I need in ~ 150 lines of code:

- blocks on subsequent calls if the cache is busy being read/written so I don't exhaust my quotas to the backend API's I'm calling.
- expires cache at midnight when the date rolls over.  This allows me to rebuild the cache programmatically during nightly low volume time instead of having cache misses impact my users sporadically throughout the day.  
- no cache in memory.  No need for wasting memory caching data, I just need to improve over 100ms API calls. ~ 1-2ms for access a local file is good enough.
- scales pretty well on the ext4 filesystem on linux - creates up to 65536 subdirectories to keep the number of files per directory, small enough that ext4 does not have a performance impact with multiple files in the same directory. 
- uses gzip compression to save space.

To use this without blowing out your filesystem, you will also need to add a cron job that periodically cleans out the cache and lock files.  For example:

```
## Delete cache files every 7 days
15 0 * * * find /var/tmp/mzcache/ -type f -mtime +7 -delete 
## Delete lock files every day
15 0 * * * find /var/tmp/mz* -type f -mtime +1 -delete
```

Of course, you will also want to set up proactive alerting on your servers to detect runaway space usage intra-day.

The point of this is that writing a simple caching routine should not be difficult and might be superior to figuring out how to use a multi-thousand line caching program like [Varnish](https://varnish-cache.org) with it's own configuration language.

Maintenance is also trivial, no need for a beefy Varnish server or servers, no need for keeping up to date with the Varnish releases and security updates, etc.  Just include in your app and deploy.

## Usage

```
package main

import (
	"github.com/sethgecko13/mzcache"
)

func cache() {
    key := "something_unique_to_cache"
    value := `data to cache
              and more`
	err := mzcache.Write(key, value)
	if err != nil {
        log.Printf("an error occurred: %s", err.Error())
	}
    result, err := mzcache.Read(key)
	if err != nil {
        log.Printf("an error occurred: %s", err.Error())
	}
    log.Printf("result from cache is: %s", result)
}
```

## Limitations

On the flip side, there are limitations:

- caches text files only.
- there is a little bit of a performance impact always converting string to bytes to write and then bytes back to string to read.
- does not do buffered reads or writes.  Every file being cached is loaded entirely into memory.  This will not scale with large files or really high volumes of small files.
- only tested on ext4 filesystem on linux and APFS on macOS, may behave differently on other filesystems.
- uses one inode per file.  So on a typical linux installation with the ext4 filesystem, you can only cache ~ 65k files before running out of inodes.  On other filesystems, or with other formatting options, this could scale beyond 65k.
- not terribly efficient on space.  Caching small gzipped files on ext4 still results in each file being 4096 bytes minimum because of the default block size.
- both read and write access to each cached file is single threaded by design to avoid race conditions.  1000 requests each accessing the same file at the same time will take access_time * 1000 to complete.  In practice this isn't a problem for my app, but might be a problem for other apps.
- cache expires at midnight.  If you want different granularity, you'll have to do it differently.

But I would argue these are limitations you need to understand anyway, it's just less obvious when you use a 3rd party cache unless you look deeply into the code or documentation.