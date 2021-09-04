#!/bin/sh
mkdir ./jpeg 2>/dev/null
tiff_file=`echo $1 | cut -f 1 -d .`.tiff
tiff2_file="out_$tiff_file"
tmp=`echo $1 | cut -f 1 -d .`.jpeg
jpeg_file="./jpeg/$tmp"
./dcraw.sh "$1" 1>/dev/null 2&>1
./tdf "$tiff_file" 1>/dev/null 2&>1
./convert.sh "$tiff2_file" "$jpeg_file" 1>/dev/null 2&>1
rm -f "$tiff_file" "$tiff2_file"
ls -l "$jpeg_file"
