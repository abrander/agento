#!/bin/bash

uptime | awk -F "[:,]" '{print "loadavg1.value" $8 "\nloadavg5.value" $9 "\nloadavg15.value" $10}'
