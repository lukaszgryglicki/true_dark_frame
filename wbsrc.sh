#!/bin/bash
# WB=filename.NEF - white balance reference file
# MULT="R G B G" - manually specify white balance, example: 0.37033 1.0 0.525732 0.996594, overrides WB
# DCARGS="-v -H 0" - manually specify dcraw args for conversion, default is "-H 1 -W -q 3 -6", other can be "-v -H 0 -q 0 -b 1.2 -g 2.1 4.2" for example
# The most imporant dcraw param is: -H [0-9] Highlight mode (0=clip, 1=unclip, 2=blend, 3+=rebuild)
# -W Don't automatically brighten the image
if [ -z "${MULT}" ]
then
  if [ -z "${WB}" ]
  then
    echo "$0: you need to specify WB source via WB=filename.NEF or specify multipliers via MULT='R B B G'"
    exit 1
  fi
  inf=`echo "${WB}" | cut -f 1 -d .`.inf
  if [ ! -f "${inf}" ]
  then
    ppm=`echo "${WB}" | cut -f 1 -d .`.ppm
    echo "generating white balance info: ${inf}"
    dcraw -v -H 1 -W -a "${WB}" 2> "${inf}" || exit 2
    rm -rf "${ppm}"
  fi
  mult=`cat "${inf}" | grep multipliers`
  mult="${mult/multipliers /}"
  export MULT="${mult}"
fi
if [ -z "${DCARGS}" ]
then
  export DCARGS="-H 1 -W -q 3 -6"
fi
echo "using RGRB multipliers: ${MULT}, dcraw args: ${DCARGS}"
for f in "$@"
do
  if [ "${f}" = "${WB}" ]
  then
    echo "$0: skipping white balance source file: ${WB}"
    continue
  fi
  tiff=`echo "${f}" | cut -f 1 -d .`.tiff
  jpeg=`echo "${f}" | cut -f 1 -d .`.jpeg
  dcraw ${DCARGS} -r ${MULT} -T "${f}" || exit 3
  convert "${tiff}" -quality 99% "${jpeg}" || exit 4
  rm -rf "${tiff}"
  RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 RLO=.2 RHI=.2 GLO=.2 GHI=.2 BLO=.2 BHI=.2 NA=1 ACM=1 Q=95 jpeg "$jpeg" || exit 5
  rm -rf "${jpeg}"
done
