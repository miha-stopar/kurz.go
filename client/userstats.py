import requests
import urllib
import urllib2
import json

data = {"id" : "3232"}
enc_data = urllib.urlencode(data)
url = "http://localhost:9999/user/"
u = urllib2.urlopen(url, enc_data)
j = u.read()
js = json.loads(j)
print js
print "--------------"
for k, v in js.iteritems():
    print k, v
