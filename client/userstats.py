import requests
import urllib
import urllib2
import json

url = "http://localhost:9999/user/"
for i in range(5):
	uid = str(101+i)
	print "user id: %s" % uid
	data = {"id" : uid}
	enc_data = urllib.urlencode(data)
	u = urllib2.urlopen(url, enc_data)
	j = u.read()
	js = json.loads(j)
	print js
	print "--------------"
