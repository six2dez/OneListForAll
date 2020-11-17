#!/bin/bash
for f in dict/*.txt;
	do cat $f >> onelistforall_all.txt
done
for f in dict/*_short.txt;
	do cat $f >> onelistforall_short_big.txt
done
duplicut onelistforall_all.txt -c -o onelistforall.txt
duplicut onelistforall_short_big.txt -c -o onelistforallshort.txt
rm -f onelistforall_*.txt
