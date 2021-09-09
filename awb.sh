#!/bin/bash
# WBRSC=white-balance-source.jpg
IJQUAL=99
if [ -z "${WBSRC}" ]
then
  size=`convert "$1" -format "%wx%h" info:`
  convert "$1" \( -clone 0 -resize 1x1! -resize $size! -modulate 100,100,0 \) \( -clone 0 -fill "gray(50%)" -colorize 100 \) -compose colorize -composite -quality "${IJQUAL}" "wb2_${1}"
else
  color=`convert "${WBSRC}" -resize 1x1! -modulate 100,100,0 -format "%[pixel:u.p{0,0}]" info:`
  convert "$1" -colorspace sRGB \( -clone 0 -fill "$color" -colorize 50% \) -compose colorize -composite -colorspace sRGB -quality "${IJQUAL}" "wb1_${1}"
fi
