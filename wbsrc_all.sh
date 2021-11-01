#!/bin/bash
for f in WB*.NEF
do
  WB="${f}" ./wbsrc.sh *.NEF
  dir="${f/.NEF/}"
  mkdir "${dir}"
  mv co_*.jpeg "${dir}"
done
