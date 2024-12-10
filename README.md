# mzcache

This is my caching library created as a response to this question on reddit:

https://www.reddit.com/r/golang/comments/1h4ubw9/should_i_roll_my_own_caching_server/

I had written this a couple of years ago because I thought it would be simpler to write my own filesystem cache that did exactly what I wanted than it would be to figure out how to write a config file to an existing cache program or library.

Overall, I think I was correct to write it myself.

This caching library does exactly what I need in ~ 150 lines of code:

- block on subsequent calls if the cache is busy being read/written.
- expire cache when the date rolls over.
- no cache in memory.
- scales on the ext4 filesystem on linux.
- uses gzip compression.

To use this without blowing out your filesystem, you will also need to add a cron job that periodically cleans out the cache and lock files.  For example:

15 0 * * * find /var/tmp/mzcache/ -type f -mtime +7 -delete
15 0 * * * find /var/tmp/mz* -type f -mtime +1 -delete

You should probably read this and use it to write your own filesystem caching routines.  The point of this is that writing a simple caching routine should not be difficult and might be superior to figuring out how to use a multi-thousand line caching program like [Varnish](https://varnish-cache.org) with it's own configuration language.
