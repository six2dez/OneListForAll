#!/bin/bash

regexes=(
    "[\!(,%]" # Ignore noisy characters
    ".{100,}" # Ignore lines with more than 100 characters (overly specific)
    "[0-9]{4,}" # Ignore lines with 4 or more consecutive digits (likely an id)
    "[0-9]{3,}$" # Ignore lines where the last 3 or more characters are digits (likely an id)
    "[a-z0-9]{32}" # Likely MD5 hash or similar
    "[0-9]+[A-Z0-9]{5,}" # Number followed by 5 or more numbers and uppercase letters (almost all noise)
    "\/.*\/.*\/.*\/.*\/.*\/.*\/" # Ignore lines more than 6 directories deep (overly specific)
    "\w{8}-\w{4}-\w{4}-\w{4}-\w{12}" # Ignore UUIDs
    "[0-9]+[a-zA-Z]+[0-9]+[a-zA-Z]+[0-9]+" # Ignore multiple numbers and letters mixed together (likley noise)
    "\.(png|jpg|jpeg|gif|svg|bmp|ttf|avif|wav|mp4|aac|ajax|css|all|)$" # Ignore low value filetypes
)

# Full list
echo "[+] Building lists..."
for f in dict/*.txt;
	do cat $f >> onelistforall_all_big.txt
done
# Short list
for f in dict/*_short.txt;
	do cat $f >> onelistforall_short_big.txt
done
echo "[+] Building done!"

# Removing buggy lines
echo "[+] Cleaning lines..."
for regex in "${regexes[@]}"; do
    cmd="cat onelistforall_all_big.txt | grep -avE '${regex}'"
    cmd="cat onelistforall_short_big.txt | grep -avE '${regex}'"
done
echo "[+] Cleaning done!"

echo "[+] Deduplication in progress..."
duplicut onelistforall_all_big.txt -c -o onelistforall.txt
duplicut onelistforall_short_big.txt -c -o onelistforallshort.txt
echo "[+] Deduplication done!"
final_lines=$(cat onelistforall.txt | wc -l)
final_lines_short=$(cat onelistforallshort.txt | wc -l)
echo "[+] onelistforall.txt has ${final_lines} lines"
echo "[+] onelistforallshort.txt has ${final_lines_short} lines"
rm -f onelistforall_*.txt
echo "[+] End"