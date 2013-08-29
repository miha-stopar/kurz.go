import requests
import urllib
import urllib2

url = "http://localhost:9999/shorten/"
for i in range(5):
	link = "https://github.com/" + str(i+1)
	if i % 3 == 0:
	    typ = "invite"
	elif i % 3 == 1:
	    typ = "share"
	else:
	    typ = "attend"
	data = {"url" : link, "eventid": str(201+i), "user" : str(101+i), "type":typ}
	enc_data = urllib.urlencode(data)
	u = urllib2.urlopen(url, enc_data)
	print u.read()
	print "-----------"
