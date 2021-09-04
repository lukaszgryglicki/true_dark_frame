#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <tiffio.h>

void error(char* s)
{
  printf("%s\n", s);
  exit(1);
}

int main(int argc, char** argv)
{
  TIFF *tif, *outtif;
  int a, i, r, c;
  uint32 w, h;
  uint16 bps, spp, phm;
  size_t scanline, bs;
  tdata_t buf;
  uint16* data;
  uint64 sums[4], cnts[4], osums[4];
  long double avg, aval, bias[4];
  char* outfn;
  if (argc < 2) error("Please provide file name");
  for (a=1;a<argc;a++)
  {
    printf("Opening: %s\n", argv[a]);
    tif = TIFFOpen(argv[a], "r");
    if (!tif) 
    {
      printf("Cannot read image: %s\n", argv[a]);
      exit(1);
    }
    TIFFGetField(tif, TIFFTAG_IMAGEWIDTH, &w);
    TIFFGetField(tif, TIFFTAG_IMAGELENGTH, &h);
    TIFFGetField(tif, TIFFTAG_BITSPERSAMPLE, &bps);
    TIFFGetField(tif, TIFFTAG_SAMPLESPERPIXEL, &spp);
    TIFFGetField(tif, TIFFTAG_PHOTOMETRIC, &phm);
    if ((w % 2) || (h % 2) || bps != 16 || spp != 1)
    {
      printf("Bad TIFF type w=%d x h=%d x bps=%d x spp=%d\n", w, h, bps, spp);
      exit(1);
    }
    scanline = TIFFScanlineSize(tif);
    /*printf("w=%d x h=%d x bps=%d x spp=%d x scanline=%zu\n", w, h, bps, spp, scanline);*/
    bs = scanline >> 1;
    buf = _TIFFmalloc(scanline);
    for (i=0;i<4;i++) sums[i] = cnts[i] = 0;
    for (r=0;r<h;r++)
    {
      TIFFReadScanline(tif, buf, r, 0);
      data=(uint16*)buf;
      for (c=0;c<bs;c++)
      {
        i = ((r % 2) << 1) + (c % 2);
        sums[i] += data[c];
        cnts[i] ++;
      }
    }
    printf("[%lu %lu %lu %lu]\n", sums[0], sums[1], sums[2], sums[3]);
    printf("[%lu %lu %lu %lu]\n", cnts[0], cnts[1], cnts[2], cnts[3]);
    avg = (long double)(sums[0] + sums[1] + sums[2] + sums[3]) / 4.;
    aval = avg / (long double)(cnts[0]);
    printf("Averaging to: %.0Lf / %.0Lf\n", avg, aval);
    for (i=0;i<4;i++) bias[i] = sums[i] / avg;
    printf("[%Lf %Lf %Lf %Lf]\n", bias[0], bias[1], bias[2], bias[3]);
    outfn = (char*)malloc(strlen(argv[a])+5);
    sprintf(outfn, "out_%s", argv[a]);
    outtif = TIFFOpen(outfn, "w");
    if (!outtif) 
    {
      printf("Cannot write image: %s\n", outfn);
      exit(1);
    }
    TIFFSetField(outtif, TIFFTAG_IMAGEWIDTH, w);
    TIFFSetField(outtif, TIFFTAG_IMAGELENGTH, h);
    TIFFSetField(outtif, TIFFTAG_BITSPERSAMPLE, bps);
    TIFFSetField(outtif, TIFFTAG_SAMPLESPERPIXEL, spp);
    TIFFSetField(outtif, TIFFTAG_PHOTOMETRIC, phm);
    for (i=0;i<4;i++) osums[i] = 0;
    for (r=0;r<h;r++)
    {
      TIFFReadScanline(tif, buf, r, 0);
      data=(uint16*)buf;
      for (c=0;c<bs;c++)
      {
        i = ((r % 2) << 1) + (c % 2);
	data[c] = (uint16)((long double)data[c] / bias[i]);
        osums[i] += data[c];
      }
      TIFFWriteScanline(outtif, (unsigned char*)data, r, 0);
    }
    printf("[%lu %lu %lu %lu]\n", osums[0], osums[1], osums[2], osums[3]);
    _TIFFfree(buf);
    TIFFClose(tif);
    TIFFClose(outtif);
    printf("Saved: %s\n", outfn);
    free((void*)outfn);
  }
}
