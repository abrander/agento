#!/bin/bash

function trap_handler()
{
    echo "FAILED: line ${1}: exit status of last command: ${2}"
    exit 1
}
trap 'trap_handler ${LINENO} $?' ERR

VERSION="0.0-$(date +"%Y%m%d-%H%M")-$(git log -n 1 --pretty="format:%h")"

go get ./...

# Package client

mkdir -p deb/DEBIAN
cat <<EOF > deb/DEBIAN/control
Package: agento-client
Version: ${VERSION}
Homepage: https://agento.org/
Section: non-free
Priority: optional
Architecture: amd64
Maintainer: Anders Brander <anders@brander.dk>
Description: Agento metric collecting agent and server
 This package contains agento-client
EOF

mkdir -p deb/etc/init
cat <<EOF > deb/etc/init/agento-client.conf
start on runlevel [2345]

respawn

setuid nobody
setgid nogroup

script
    exec /usr/sbin/agento-client
end script
EOF

(cd client ; go build .)
mkdir -p deb/usr/sbin/
cp -a client/client deb/usr/sbin/agento-client

dpkg-deb --build deb $(pwd)
rm -rf deb

# Package server

mkdir -p deb/DEBIAN
cat <<EOF > deb/DEBIAN/control
Package: agento-server
Version: ${VERSION}
Homepage: https://agento.org/
Section: non-free
Priority: optional
Architecture: amd64
Maintainer: Anders Brander <anders@brander.dk>
Description: Agento metric collecting agent and server
 This package contains agento-server
EOF

mkdir -p deb/etc/init
cat <<EOF > deb/etc/init/agento-server.conf
start on runlevel [2345]

respawn

setuid nobody
setgid nogroup

script
    exec /usr/sbin/agento-server
end script
EOF

(cd server ; go build .)
mkdir -p deb/usr/sbin/
cp -a server/server deb/usr/sbin/agento-server

dpkg-deb --build deb $(pwd)
rm -rf deb
