#!/bin/bash
# WBRSC=white-balance-source.jpg
tiff=`echo $1 | cut -f 1 -d .`.tiff
jpeg=`echo $1 | cut -f 1 -d .`.jpeg
png=`echo $1 | cut -f 1 -d .`.png
function cleanup {
  rm -rf "$tiff" "$jpeg" "$png" 1>/dev/null 2>/dev/null
}
trap cleanup EXIT
#dcraw -v -H 0 -A 200 200 5600 3600 -T -q 3 "$1" || exit 1
# -H 1, 5+ looks best for me
dcraw -v -H 1 -a -T -q 3 "$1" || exit 1
if [ -z "${WBSRC}" ]
then
  convert "$tiff" -quality 98% "$jpeg" || exit 2
else
  convert "$tiff" "$png" || exit 3
  color=`convert "${WBSRC}" -resize 1x1! -modulate 100,100,0 -format "%[pixel:u.p{0,0}]" info:`
  convert "$png" -colorspace sRGB \( -clone 0 -fill "$color" -colorize 50% \) -compose colorize -composite -colorspace sRGB -quality 98% "$jpeg" || exit 4
fi
RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 RLO=.2 RHI=.2 GLO=.2 GHI=.2 BLO=.2 BHI=.2 NA=1 Q=95 jpeg "$jpeg" || exit 5
