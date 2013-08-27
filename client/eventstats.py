import requests
import urllib
import urllib2
import json

url = "https://github.com/miha-stopar"
data = {"url" : url}
enc_data = urllib.urlencode(data)
url = "http://localhost:9999/event/"
u = urllib2.urlopen(url, enc_data)
j = u.read()
js = json.loads(j)
print js
