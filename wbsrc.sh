#!/bin/bash
# WB=filename.NEF - white balance reference file
# NOLG=1 - do not use lukaszgryglicki's "jpeg/jpegbw" tools - use ImageMagick's "convert" instead
# MULT="R G B G" - manually specify white balance, example: 0.37033 1.0 0.525732 0.996594, overrides WB
# DCWBARGS="-v -H 0 -W -a" - manually specify dcraw args for white balance source conversion, default is "-v -H 1 -W -a"
# DCARGS="-v -H 1" - manually specify dcraw args for conversion, default is "-v -H 0 -W -q 3 -6", other can be "-H 1 -q 0 -b 1.2 -g 2.1 4.2" for example
# The most imporant dcraw param is: -H [0-9] Highlight mode (0=clip, 1=unclip, 2=blend, 3+=rebuild)
# -W Don't automatically brighten the image
# LO=0-100 - set LO enhance cutoff - default 0.5 (percent)
# HI=0-100 - set HI enhance cutoff - default 0.5 (percent)
# NODEL=1 - do not remove intermediate files
# NOJPEGARGS - do ot use built in jpeg command args (assume they are set as env)
if [ -z "${MULT}" ]
then
  if [ -z "${WB}" ]
  then
    echo "$0: you need to specify WB source via WB=filename.NEF or specify multipliers via MULT='R G B G'"
    exit 1
  fi
  inf="${WB/.NEF/.inf}"
  if [ ! -f "${inf}" ]
  then
    ppm="${WB/.NEF/.ppm}"
    if [ -z "${DCWBARGS}" ]
    then
      export DCWBARGS="-v -H 1 -W -a"
      echo "using default white balance source dcraw args: ${DCWBARGS}"
    fi
    echo "generating white balance info: ${inf}"
    dcraw ${DCWBARGS} "${WB}" 2> "${inf}" || exit 2
    if [ -z "$NODEL" ]
    then
      rm -rf "${ppm}"
    fi
  fi
  mult=`cat "${inf}" | grep multipliers`
  mult="${mult/multipliers /}"
  export MULT="${mult}"
fi
if [ -z "${DCARGS}" ]
then
  export DCARGS="-v -H 0 -W -q 3 -6"
fi
if [ -z "$LO" ]
then
  export LO="0.5"
fi
if [ -z "$HI" ]
then
  export HI="0.5"
fi
echo "using RGRB multipliers: ${MULT}, dcraw args: ${DCARGS}, lo/hi=$LO/$HI"
for f in "$@"
do
  if [[ $f != *.NEF ]]
  then
    echo "$0: ${f} is not a NEF file, skipping"
    continue
  fi
  if [ "${f}" = "${WB}" ]
  then
    echo "$0: skipping white balance source file: ${WB}"
    continue
  fi
  echo "processing ${f}"
  tiff="${f/.NEF/.tiff}"
  jpeg="${f/.NEF/.jpeg}"
  dcraw ${DCARGS} -r ${MULT} -T "${f}" || exit 3
  if [ -z "${NOLG}" ]
  then
    convert "${tiff}" -enhance -quality 99% "${jpeg}" || exit 4
    if [ -z "$NODEL" ]
    then
      rm -rf "${tiff}"
    fi
    if [ -z "$NOJPEGARGS" ]
    then
      RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 RLO=$LO RHI=$HI GLO=$LO GHI=$HI BLO=$LO BHI=$HI NA=1 ACM=1 Q=96 jpeg "${jpeg}" || exit 5
    else
      jpeg "${jpeg}" || exit 6
    fi
    if [ -z "$NODEL" ]
    then
      rm -rf "${jpeg}"
    fi
    echo "saved co_${jpeg}"
  else
    convert "${tiff}" -auto-level -enhance -quality 95% "${jpeg}" || exit 7
    if [ -z "$NODEL" ]
    then
      rm -rf "${tiff}"
    fi
    echo "saved ${jpeg}"
  fi
done
