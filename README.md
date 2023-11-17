This repo contains a Go port of [gunzip](https://github.com/TechHara/gunzip) for decompression of `.gz` files.

# Benchmark
Below shows runtime comparison with the go standard library implementation (compress/gzip).
On Linux x64 systems, there is significant performance regression if run without explicitly limiting CPU affinity with `taskset`. See run commands below.

## Decompression of linux.tar.gz (Linux x64)
|  # Gorutines | compress/gzip  | This  |
|:-:|:-:|:-:|
| 1 | 3.68 | 3.52 |
| 2 | | 3.51 |

## Decompression of linux.tar.gz (macOS arm64)
|  # Gorutines | compress/gzip  | This  |
|:-:|:-:|:-:|
| 1 | 4.12 | 4.71 |
| 2 | | 4.42 |


# Build
```sh
$ go build
```

# Run
```sh
# single gorutine
$ ./gunzip < compressed.gz > decompressed
# On Linux x64, run with explicit CPU affinity
$ taskset -c 0 ./gunzip < compressed.gz > decompressed

# two gorutines
$ ./gunzip -t < compressed.gz > decompressed
# On Linux x64, run with explicit CPU affinity
$ taskset -c 0,2 ./gunzip -t < compressed.gz > decompressed
```
