import requests
import urllib
import urllib2

data = {"id" : "3232"}
enc_data = urllib.urlencode(data)
url = "http://localhost:9999/user/"
u = urllib2.urlopen(url, enc_data)
print u.read()
