import requests
import urllib
import urllib2

url = "http://192.168.1.13:9999/resolve/"
data = {"url" : url, "short": "BY2B5"}

enc_data = urllib.urlencode(data)
u = urllib2.urlopen(url, enc_data)
print u.read()
print "-----------"
