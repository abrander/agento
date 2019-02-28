#!/bin/bash

function trap_handler()
{
    echo "FAILED: line ${1}: exit status of last command: ${2}"
    exit 1
}
trap 'trap_handler ${LINENO} $?' ERR

VERSION="0.0-$(date +"%Y%m%d-%H%M")-$(git log -n 1 --pretty="format:%h")"

go get -u ./...

# Package client

mkdir -p deb/DEBIAN
cat <<EOF > deb/DEBIAN/control
Package: agento
Version: ${VERSION}
Homepage: https://agento.org/
Section: non-free
Priority: optional
Architecture: amd64
Maintainer: Anders Brander <anders@brander.dk>
Description: Agento metric collecting agent and server
 This package contains agento, an metrics collecting agent and server
EOF

mkdir -p deb/etc/init
cat <<EOF > deb/etc/init/agento.conf
start on runlevel [2345]

respawn

setuid nobody
setgid nogroup

script
    if [ -r /etc/default/agento ]; then
        . /etc/default/agento
        export AGENTO_SECRET AGENTO_SERVER_URL AGENTO_INFLUXDB_URL AGENTO_MONGO_URL
    fi

    exec /usr/sbin/agento run
end script
EOF

go build .
mkdir -p deb/usr/sbin/
cp -a agento deb/usr/sbin/agento

cat <<EOF > deb/DEBIAN/postinst
#!/bin/sh

exec setcap CAP_NET_RAW=ep /usr/sbin/agento
EOF
chmod 755 deb/DEBIAN/postinst

dpkg-deb --build deb $(pwd)
rm -rf deb
