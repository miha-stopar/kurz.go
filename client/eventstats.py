import requests
import urllib
import urllib2
import json

for i in range(5):
	eid = str(201 + i)
	print "event id: %s" % eid
	data = {"eventid" :eid}
	enc_data = urllib.urlencode(data)
	eurl = "http://192.168.1.13:9999/event/"
	u = urllib2.urlopen(eurl, enc_data)
	j = u.read()
	js = json.loads(j)
	print js
	print "--------------"
