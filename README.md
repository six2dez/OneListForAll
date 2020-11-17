# OneListForAll
**Rockyou for web fuzzing**

**V2 released!** Now you can build your own wordlists with the same method and this release includes a **short** wordlist. Base wordlists provided in /dict folder. **See Method 3**

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

## Wordlists content

In the fields that both lists coincide, the short one has the content but in less quantity, only the most relevant.

Both lists have:

- First slash (/) removed, lines that have it is on purpose.
- Removed special chars or crash chars such as `' sqlis, xss, etc
- Trimmed trailing whitespaces
- Removed comments (lines starting with #)

| Year               | Short              | Full               |
| ----               | -----------------  | -------------------|
| Size               |                5M  |               180M |
| Lines              |            344644  |            9117326 |
| Extension specific |            &check; | :heavy_check_mark: |
| Config files       |           &check;  | :heavy_check_mark: |
| Admin panels       |           &check;  | :heavy_check_mark: |
| Dotfiles           |           &check;  | :heavy_check_mark: |
| Backup files/folders |           &check;  | :heavy_check_mark: |
| LFI                |           &check;  | :heavy_check_mark: |
| Multilanguage dicts  |           &check;  | :heavy_check_mark: |
| Extension specific |           &check;  | :heavy_check_mark: |
| CMS specific |           &check;  | :heavy_check_mark: |
| Robots Disallowed |           &check;  | :heavy_check_mark: |
| Software specific  |           &check;  | :heavy_check_mark: |
|          Usernames |           &cross;  | :heavy_check_mark: |
|          Words     |           &cross;  | :heavy_check_mark: |
|     Subdomains     |           &cross;  | :heavy_check_mark: |

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
- [bruteforce-list](https://github.com/random-robbie/bruteforce-lists)

Feel free to contribute, PR are welcomed.

You can support this work buying me a coffee:

[<img src="https://www.buymeacoffee.com/assets/img/guidelines/bmc-coffee.gif">](https://www.buymeacoffee.com/six2dez)
