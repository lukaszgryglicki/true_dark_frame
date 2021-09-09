#!/bin/bash
# FPS=29.97
# JQUAL=90
# IJQUAL=99
# VQ=
if [ -z "${FPS}" ]
then
  result=`ffprobe -v error -select_streams v -of default=noprint_wrappers=1:nokey=1 -show_entries stream=r_frame_rate "${1}"`
  FPS=`echo "scale=3; ${result}" | bc`
  if [ -z "${FPS}" ]
  then
    FPS="29.970"
  else
    echo "detected fps=${FPS}"
  fi
fi
if [ -z "${JQUAL}" ]
then
  JQUAL=90
fi
if [ -z "${IJQUAL}" ]
then
  IJQUAL=99
fi
echo "jpeg quality: ${IJQUAL}/${JQUAL}"
if [ -z "${VQ}" ]
then
  VQ=20
fi
echo "video quality: ${VQ}"
root=`echo $1 | cut -f 1 -d .`
ffmpeg -i "$1" -qmin 1 -qmax "${VQ}" "${root}_%06d.png" || exit 1
ffmpeg -i "$1" -vn -acodec aac -ac 2 -ar 48000 -f mp4 -y "${root}.aac" || exit 2
for f in ${root}_*.png
do
  if [ -z "$size" ]
  then
    size=`convert "$f" -format "%wx%h" info:`
  fi
  jf=`echo $f | cut -f 1 -d .`.jpeg
  convert "${f}" \
  \( -clone 0 -resize 1x1! -resize $size! -modulate 100,100,0 \) \
  \( -clone 0 -fill "gray(50%)" -colorize 100 \) \
  -compose colorize -composite -quality "${IJQUAL}" "${jf}" || exit 3
  Q="${JQUAL}" jpeg.sh "${jf}" || exit 4
  rm -f "${f}" "${jf}" || exit 5
done
ffmpeg -framerate "${FPS}" -i "co_${root}_%06d.jpeg" -r "${FPS}" -i "${root}.aac" -s "${size}" -vcodec h264 -mbd 2 -preset slower -crf "${VQ}" -shortest -y "${root}.mp4" || exit 6
rm -f co_${root}_*.jpeg || exit 7
echo "OK: ${root}.mp4"
