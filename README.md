# OneListForAll
**Rockyou for web fuzzing**

This is a project to generate huge wordlists for web fuzzing, if you just want to fuzz with a good wordlist use the file [onelistforallmicro.txt](https://github.com/six2dez/OneListForAll/blob/main/onelistforallmicro.txt).

## Usage

### Method 1

1. Go to [releases](https://github.com/six2dez/OneListForAll/releases) and download

2. Fuzz with the best tool [ffuf](https://github.com/ffuf/ffuf) :)
```bash
ffuf -c -w onelistforall.txt -u [target.com]/FUZZ
```

### Method 2

1. Git clone and extract:
```bash
git clone https://github.com/six2dez/OneListForAll && cd OneListForAll
7z x onelistforall.7z.001
```
2. Fuzz with the best tool [ffuf](https://github.com/ffuf/ffuf) :)
```bash
ffuf -c -w onelistforall.txt -u [target.com]/FUZZ
```

### Method 3

**Build your own wordlists!**

1. Add your wordlists to dict/ folder with suffix **_short.txt** for short wordlist and **_long.txt** for the full wordlist.

2. Run ./olfa.sh (olfa -> One List For All) and you will have onelistforall.txt file and onelistforallshort.txt.

3. Fuzz with the best tool [ffuf](https://github.com/ffuf/ffuf) :)
```bash
ffuf -c -w onelistforall.txt -u [target.com]/FUZZ
```

## Wordlists summary

- **onelistforallmicro.txt** manally crafted wordlist for low hanging fruits: 18109 lines, 298K
- **onelistforallshort.txt** a shortened version, it also contains a lot of things, but in a more affordable way: 822115 lines, 12M
- **onelistforall.txt** basically everything, launch it and go to sleep. 55684781 lines, 1.1G

## Sources

This is a wordlists project for fuzzing purposes made from the best word lists currently available,merged and deduplicated later with [duplicut](https://github.com/nil0x42/duplicut), adding cleaner from [BonJarber](https://github.com/BonJarber/SecUtils/tree/master/clean_wordlist). The lists used have been selected from these repositories:

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
- [bruteforce-list](https://github.com/random-robbie/bruteforce-lists)
- [assetnote](https://wordlists.assetnote.io/)
- [brutas](https://github.com/tasooshi/brutas)
- [werdlists](https://github.com/decal/werdlists)

Feel free to contribute, PR are welcomed.