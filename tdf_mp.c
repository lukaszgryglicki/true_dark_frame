#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <tiffio.h>
#include <pthread.h>
#include <math.h>

void error(char* s)
{
  printf("%s\n", s);
  exit(1);
}

uint8 mapper(int i, int n)
{
    float f, r;
    uint8 rv;
    f = (float)i / (float)n;
    if (f <= 0.5) r = .5-(.5-f)*(.5-f)*2.;
    else r = .5+(f-.5)*(f-.5)*2.;
    rv = (uint8)(r*255.9);
/*    printf("(%d,%d) --> %f,%f,%d\n", i, n, f, r, rv);*/
    return rv;
}

void compute_occurence_table(uint8* pre)
{
    int i, n;
    uint16 *occ;
    n = 0;
    for (i=0;i<0x10000;i++) if (pre[i]) n ++;
    printf("Distinct 16bit values: %d\n", n);
    occ = (uint16*)malloc(n*sizeof(uint16));
    n = 0;
    for (i=0;i<0x10000;i++) if (pre[i]) occ[n++] = i;
    for (i=0;i<n;i++) pre[occ[i]] = mapper(i, n);

}

void * work(void* varg)
{
    TIFF *tif, *outtif;
    int i, r, c;
    uint32 w, h;
    uint16 bps, spp;
    size_t scanline, bs;
    tdata_t buf, buf2;
    uint16 *data;
    uint8 *data2;
    uint64 sums[4], cnts[4], osums[4];
    uint8* pre;
    long double avg, aval, bias[4];
    char *outfn, *filename;
    filename = (char*)varg;
    pre = (uint8*)malloc(0x10000);
    memset((void*)pre, 0, 0x10000);
    printf("Opening: %s\n", filename);
    tif = TIFFOpen(filename, "r");
    if (!tif) 
    {
      printf("Cannot read image: %s\n", filename);
      exit(1);
    }
    TIFFGetField(tif, TIFFTAG_IMAGEWIDTH, &w);
    TIFFGetField(tif, TIFFTAG_IMAGELENGTH, &h);
    TIFFGetField(tif, TIFFTAG_BITSPERSAMPLE, &bps);
    TIFFGetField(tif, TIFFTAG_SAMPLESPERPIXEL, &spp);
    if ((w % 2) || (h % 2) || bps != 16 || spp != 1)
    {
      printf("Bad TIFF type w=%d x h=%d x bps=%d x spp=%d\n", w, h, bps, spp);
      exit(1);
    }
    scanline = TIFFScanlineSize(tif);
    /*printf("w=%d x h=%d x bps=%d x spp=%d x scanline=%zu\n", w, h, bps, spp, scanline);*/
    bs = scanline >> 1;
    buf = _TIFFmalloc(scanline);
    buf2 = _TIFFmalloc(bs);
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
    outfn = (char*)malloc(strlen(filename)+5);
    /* 16bit output*/
    sprintf(outfn, "out_%s", filename);
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
    for (i=0;i<4;i++) osums[i] = 0;
    for (r=0;r<h;r++)
    {
      TIFFReadScanline(tif, buf, r, 0);
      data=(uint16*)buf;
      for (c=0;c<bs;c++)
      {
        i = ((r % 2) << 1) + (c % 2);
	data[c] = (uint16)((long double)data[c] / bias[i]);
	pre[data[c]] = 1;
        osums[i] += data[c];
      }
      TIFFWriteScanline(outtif, (unsigned char*)data, r, 0);
    }
    printf("[%lu %lu %lu %lu]\n", osums[0], osums[1], osums[2], osums[3]);
    compute_occurence_table(pre);
    TIFFClose(outtif);
    printf("Saved: %s\n", outfn);
    /* 8bit occurence based output*/
    sprintf(outfn, "opt_%s", filename);
    outtif = TIFFOpen(outfn, "w");
    if (!outtif) 
    {
      printf("Cannot write image: %s\n", outfn);
      exit(1);
    }
    TIFFSetField(outtif, TIFFTAG_IMAGEWIDTH, w);
    TIFFSetField(outtif, TIFFTAG_IMAGELENGTH, h);
    TIFFSetField(outtif, TIFFTAG_BITSPERSAMPLE, 8);
    TIFFSetField(outtif, TIFFTAG_SAMPLESPERPIXEL, spp);
    for (r=0;r<h;r++)
    {
      TIFFReadScanline(tif, buf, r, 0);
      data=(uint16*)buf;
      data2=(uint8*)buf2;
      for (c=0;c<bs;c++)
      {
        i = ((r % 2) << 1) + (c % 2);
	data[c] = (uint16)((long double)data[c] / bias[i]);
	data2[c] = pre[data[c]];
      }
      TIFFWriteScanline(outtif, (unsigned char*)data2, r, 0);
    }
    TIFFClose(outtif);
    printf("Saved: %s\n", outfn);
    free((void*)outfn);
    _TIFFfree(buf);
    TIFFClose(tif);
    return NULL;
}

int main(int argc, char** argv)
{
  int nt, a;
  pthread_t *threads;
  if (argc < 3) error("Please provide <threads_num> file name(s)");
  nt = atoi(argv[1]);
  threads = (pthread_t*)malloc(nt*sizeof(pthread_t));
  if (nt < 1) error("Plese provide number of threads >= 1");
  if (argc != nt + 2) error("Please provide the same numbe rof files to process as threads number");
  for (a=2;a<argc;a++)
  {
    if (pthread_create(&threads[a-2], NULL, work, (void*)(argv[a]))) error("Failed to create thread");
  }
  for (a=0;a<nt;a++) pthread_join(threads[a], NULL);
  free((void*)threads);
  return 0;
}
