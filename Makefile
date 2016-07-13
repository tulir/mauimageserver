build:
	go build -o mauimageserver

package-prep: build
	mv mauimageserver package/usr/bin/
	cp image.html package/etc/mis/
	cp config.json package/etc/mis/

package: package-prep
	dpkg-deb --build package mauimageserver.deb > /dev/null

clean:
	rm -f mauimageserver mauimageserver.deb package/usr/bin/mauimageserver package/etc/mis/image.html package/etc/mis/config.json
