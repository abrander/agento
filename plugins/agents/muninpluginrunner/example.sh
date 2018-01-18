#!/bin/bash

set -euo pipefail

uptime | grep -ohe 'load average[s:][: ].*' | awk -F '[ ,]' '{print "loadavg1.value " $3 "\nloadavg5.value " $5 "\nloadavg15.value " $7}'
