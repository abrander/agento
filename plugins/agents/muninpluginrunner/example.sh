#!/bin/bash

if [ -n "${1}" ]; then
	echo "testing 1234.1234"
fi

uptime | awk -F "[:,]" '{print "loadavg1.value" $8 "\nloadavg5.value" $9 "\nloadavg15.value" $10}'
