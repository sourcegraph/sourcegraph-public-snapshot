#!/bin/sh

case $1 in
	drfront)
		./bin/madge -i /tmp/drfront.png -f amd -x '^test$|^globals$|^app\.build$|^tests|^text\!' ~/Sites/com.aptoma.drfront/beta/drfront/public/js/modules
		;;
	nau)
		./bin/madge -i /tmp/nau.png -f cjs ~/Code/com.aptoma.drfront.nau/app ~/Code/com.aptoma.drfront.nau/lib
		;;
	vgpluss)
		./bin/madge -i /tmp/vgpluss.png -f amd ~/Code/no.vg.pluss/public/js/modules
		;;
esac