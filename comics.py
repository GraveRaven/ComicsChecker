import urllib.request, urllib.parse
import webbrowser
from multiprocessing.pool import ThreadPool

class ComicEntry:
        def __init__(self):
                self.url = ""
                self.match = ""
                self.old = ""

        def __init__(self, url, match, old):
                self.url = url
                self.match = match
                self.old = old

def CheckComic(comic):
      
        try:
                resp = opener.open(comic.url)
        except BaseException as e:
                print("ERROR: %s" % str(e))
        else:
                for line in resp:
                        line = line.decode("utf-8").strip()
                        if(comic.match in line):
                                if(line != comic.old):
                                        comic.old = line.strip()
                                        webbrowser.open(comic.url, new=0, autoraise=False)
                                break
        return comic

comics = {}
headers = { 'User-Agent' : 'Mozilla/5.0 (Windows NT 6.1; WOW64; rv:23.0) Gecko/20100101 Firefox/23.0' }
oldmatch = {}

opener = urllib.request.build_opener()
opener.addheaders = headers.items()

file = open("comic.check", "r")

todo = 1
url = ""
match = ""
old = ""
comics = []
for line in file:
	if(todo == 1):
		url = line.rstrip()
		todo = 2
	elif(todo == 2):
		match = line.rstrip()
		todo = 3
	elif(todo == 3):
		old = line.rstrip()
		todo = 1
		comics.append(ComicEntry(url, match, old))


file.close()
if(todo != 1):
	print("ERROR PARSING CONFIG")
	exit(1)

pool = ThreadPool(processes = 20)
ret = pool.map(CheckComic, comics)

file = open("comic.check", "w")
for comic in ret:
	file.write(comic.url + "\n")
	file.write(comic.match + "\n")
	file.write(comic.old + "\n")
file.close()
