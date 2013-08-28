import requests
import urllib
import urllib2
import json

for i in range(5):
	url = "https://github.com/%s" % str(i+1)
	print "event url: %s" % url
	data = {"url" : url}
	enc_data = urllib.urlencode(data)
	eurl = "http://localhost:9999/event/"
	u = urllib2.urlopen(eurl, enc_data)
	j = u.read()
	js = json.loads(j)
	print js
	print "--------------"
