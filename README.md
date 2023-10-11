This repo contains a Go port of [`gunzip`](https://github.com/TechHara/gunzip) for decompression of `.gz` files.

# Benchmark
Below shows runtime comparison with the go standard library implementation (compress/gzip). `GOMAXPROCS` is set equal to #Gorutines. Strangely, setting `GOMAXPROCS` to more than 1 in x64 causes significant performance regression.

## Decompression of linux.tar.gz (Linux x64)
|  # Gorutines | compress/gzip  | This  |
|:-:|:-:|:-:|
| 1 | 3.71 | 3.99 |
| 2 | | 6.51 |

## Decompression of linux.tar.gz (macOS arm64)
|  # Gorutines | compress/gzip  | This  |
|:-:|:-:|:-:|
| 1 | 4.12 | 4.52 |
| 2 | | 4.12 |


# Build
```sh
$ go build
```

# Run
```sh
# single gorutine
$ GOMAXPROCS=1 ./gunzip < compressed.gz > decompressed

# two gorutines
$ GOMAXPROCS=2 ./gunzip -t < compressed.gz > decompressed
```
