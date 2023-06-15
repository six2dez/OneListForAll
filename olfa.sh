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
    "[0-9]+[a-zA-Z]+[0-9]+[a-zA-Z]+[0-9]+" # Ignore multiple numbers and letters mixed together (likely noise)
    "\.(png|jpg|jpeg|gif|svg|bmp|ttf|avif|wav|mp4|aac|ajax|css|all|)$" # Ignore low value filetypes
)

# Full list
echo "[+] Building lists..."
for f in dict/*.txt; do
    cat "$f" >> onelistforall_all_tmp.txt
done
# Short list
for f in dict/*_short.txt; do
    cat "$f" >> onelistforall_short_tmp.txt
done
echo "[+] Building done!"

cmd1="cat onelistforall_all_tmp.txt"
cmd2="cat onelistforall_short_tmp.txt"
# Removing buggy lines
echo "[+] Cleaning lines..."
for regex in "${regexes[@]}"; do
    cmd1="$cmd1 | grep -avE '${regex}'"
    cmd2="$cmd2 | grep -avE '${regex}'"
done

cmd1="${cmd1} > onelistforall_big.txt"
cmd2="${cmd2} > onelistforall_short.txt"

eval "$cmd1"
eval "$cmd2"

echo "[+] Cleaning done!"

echo "[+] Deduplication in progress..."
sort onelistforall_big.txt | uniq -u > onelistforall.txt
sort onelistforall_short.txt | uniq -u > onelistforallshort.txt
echo "[+] Deduplication done!"
final_lines=$(wc -l < onelistforall.txt)
final_lines_short=$(wc -l < onelistforallshort.txt)
echo "[+] onelistforall.txt has ${final_lines} lines"
echo "[+] onelistforallshort.txt has ${final_lines_short} lines"
rm -f onelistforall_*.txt
echo "[+] End"
