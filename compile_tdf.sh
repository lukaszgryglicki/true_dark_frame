cc -Wall -ansi -pedantic -O3 -I/usr/local/include -L/usr/local/lib -o tdf tdf.c -ltiff
cc -pthread -Wall -ansi -pedantic -O3 -I/usr/local/include -L/usr/local/lib -o tdf_mp tdf_mp.c -ltiff -lm
