#!/bin/bash
# FPS=60
if [ -z "${FPS}" ]
then
  FPS="29.970"
fi
root=`echo $1 | cut -f 1 -d .`
#ffmpeg -i "$1" -vcodec h264 -mbd 2 -preset slower -crf 19 -acodec aac -ac 1 -ar 22050 -f mp4 "$out"
#ffmpeg -i _FSC3138.MOV '%04d.png'
ffmpeg -i "$1" "${root}_%06d.png" || exit 1
ffmpeg -i "$1" -vn -acodec aac -ac 2 -ar 48000 -f mp4 "${root}.aac" || exit 2
for f in ${root}_*.png
do
  if [ -z "$size" ]
  then
    size=`convert "$f" -format "%wx%h" info:`
  fi
  convert "${f}" \
  \( -clone 0 -resize 1x1! -resize $size! -modulate 100,100,0 \) \
  \( -clone 0 -fill "gray(50%)" -colorize 100 \) \
  -compose colorize -composite "wb_${f}" || exit 3
  rm -f "${f}" || exit 4
done
ffmpeg -i "wb_${root}_%06d.png" -i "${root}.aac" -r "${FPS}" -s "${size}" -vcodec h264 -mbd 2 -preset slower -crf 19 -shortest "${root}.mp4" || exit 5
rm -f "wb_${root}_%06d.png" || exit 6
echo "OK: ${root}.mp4"
