# Options

- `DEBUG=n` - debug mode.
- `OUTPUT=1` - sub commands output.
- `KEEP=1` - keep intermediate files.
- `VQ=1-31, default 20` - video quantizer setting.
- `FPS=29.97, default autodetect` - frames per second.
- `N_CPUS=16, default autodetect` - number of threads to use.
- `IJQUAL=99, default 99` - internal/intermediate jpeg files quality.
- `JQUAL=90, default 90` - output jpeg quality (frames).
- `JPEG_NO_DEFAULT=1` - no default jpeg command params, which are: `RR=1 RG=0 RB=0 GR=0 GG=1 GB=0 BR=0 BG=0 BB=1 RLO=.3 RHI=.3 GLO=.3 GHI=.3 BLO=.3 BHI=.3 NA=1`.
- `NO_JPEG=1` - don't use jpeg tool.
- `NO_CONVERT=1` - it will still use convert for png -> jpeg conversion.
- `WBRSC=white-balance-source.jpg` - use a given file jpeg/png for white balance source
- `ACM=1` - do not streat all colors independently, stretch while keepingh colors balance (ACM - all colors mode).


# To use custom white balance as returned by DCRAW

- `dcraw` will return `multipliers 0.690377 0.818478 1.000000 0.818295` when run with `-v -a` arguments.
- Use: `JPEG_NO_DEFAULT=1 NO_CONVERT=1 RR=0.690377 RG=0 RB=0 GR=0 GG=0.8183865 GB=0 BR=0 BG=0 BB=1 RLO=.3 RHI=.3 GLO=.3 GHI=.3 BLO=.3 BHI=.3 NA=1 ACM=1 awbmov *.MOV`.
