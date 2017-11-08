cp config.json myplugin/
docker plugin rm -f  suryagaddipati/cachedriver 
docker plugin create suryagaddipati/cachedriver myplugin/ 
docker plugin push suryagaddipati/cachedriver
