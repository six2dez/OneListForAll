# OneListForAll
**Rockyou for web fuzzing**

## Usage

1. Git clone and extract:
```bash
git clone https://github.com/six2dez/OneListForAll && cd OneListForAll
7z x onelistforall.7z.001
```
2. Fuzz with the best tool [ffuf](https://github.com/ffuf/ffuf) :)
```bash
ffuf -c -mc all -ac -w onelistforall.txt -u [target.com]/FUZZ
```

## Sources

This is a wordlist for fuzzing purposes made from the best wordlists currently available, lowercased and deduplicated later with [duplicut](https://github.com/nil0x42/duplicut). The lists used have been some selected within these repositories:

- [fuzzdb](https://github.com/fuzzdb-project/fuzzdb)
- [SecLists](https://github.com/danielmiessler/SecLists)
- [xmendez](https://github.com/xmendez/wfuzz)
- [minimaxir](https://github.com/minimaxir/big-list-of-naughty-strings)
- [TheRook](https://github.com/TheRook/subbrute)
- [danielmiessler](https://github.com/danielmiessler]/RobotsDisallowed)
- [swisskyrepo](https://github.com/swisskyrepo/PayloadsAllTheThings)
- [1N3](https://github.com/1N3/IntruderPayloads)
- [cujanovic](https://github.com/cujanovic)
- [lavalamp](https://github.com/lavalamp-/password-lists)
- [ics-default](https://github.com/arnaudsoullie/ics-default-passwords)
- [jeanphorn](https://github.com/jeanphorn/wordlist)
- [j3ers3](https://github.com/j3ers3/PassList)
- [nyxxxie](https://github.com/nyxxxie/awesome-default-passwords)
- [dirbuster](https://www.owasp.org/index.php/DirBuster)
- [dotdotpwn](https://github.com/wireghoul/dotdotpwn)
- [hackerone_wordlist](https://github.com/xyele/hackerone_wordlist)
- [commonspeak2](https://github.com/assetnote/commonspeak2-wordlists)
- [bruteforce-list](https://github.com/assetnote/commonspeak2-wordlists)

Feel free to contribute, PR are welcomed.

You can support this work buying me a coffee:

[<img src="https://www.buymeacoffee.com/assets/img/guidelines/bmc-coffee.gif">](https://www.buymeacoffee.com/six2dez)
