#!/bin/sh
convert "$1" -type grayscale -linear-stretch 0.5%x2% -quality 95% -depth 8 "$2"
