all:
	( cd mmark; go build )

draft2:	mmark2rfc2.txt mmark/mmark

mmark2rfc2.txt: mmark2rfc.md mmark/mmark
	./mmark/mmark -xml2 -page mmark2rfc.md > x.xml && xml2rfc --text x.xml && rm x.xml && mv x.txt mmark2rfc2.txt

mmark2rfc2.xml: mmark2rfc.md mmark/mmark
	./mmark/mmark -xml2 -page mmark2rfc.md > mmark2rfc2.xml


draft3:	mmark2rfc3.xml mmark/mmark

mmark2rfc3.xml: mmark2rfc.md mmark/mmark
	./mmark/mmark -xml -page mmark2rfc.md > mmark2rfc3.xml

.PHONY:
clean:
	rm -f mmark2rfc2.xml mmark2rfc3.xml mmark2rfc2.txt x.xml

.PHONY:
validate: mmark2rfc3.xml
	xmllint --xinclude mmark2rfc3.xml | jing -c xml2rfcv3.rnc /dev/stdin
