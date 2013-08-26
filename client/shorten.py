import requests
import urllib
import urllib2


link = "https://github.com/miha-stopar"
data = {"url" : link, "user" : "3232", "type":"invite"}
enc_data = urllib.urlencode(data)
url = "http://localhost:9999/shorten/"
u = urllib2.urlopen(url, enc_data)
print u.read()
