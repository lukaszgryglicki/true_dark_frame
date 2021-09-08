#!/bin/bash
tiff=`echo $1 | cut -f 1 -d .`.tiff
jpeg=`echo $1 | cut -f 1 -d .`.jpeg
function cleanup {
  rm -rf "$tiff" "$jpeg" 1>/dev/null 2>/dev/null
}
trap cleanup EXIT
#dcraw -v -H 0 -A 200 200 5600 3600 -T -q 3 "$1" || exit 1
dcraw -v -H 0 -a -T -q 3 "$1" || exit 1
convert "$tiff" -quality 98% "$jpeg" || exit 2
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 RLO=.2 RHI=.2 GLO=.2 GHI=.2 BLO=.2 BHI=.2 NA=1 Q=95 jpeg "$jpeg" || exit 3
