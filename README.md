| Command                        | CPU         | Benchmark                 | Iterations   | Time per Operation | Allocations per Operation |
|--------------------------------|-------------|---------------------------|--------------|--------------------|---------------------------|
| `GOMAXPROCS=1 -benchtime=10s`  | Apple M3 Pro | CalculateStdDevAndMean    | 1000000000   | 6.082 ns/op        | 0 B/op                    |
|                                |             | CalculateSizeZScore       | 295832040    | 39.77 ns/op        | 0 B/op                    |
| `GOMAXPROCS=1`                 | Apple M3 Pro | CalculateStdDevAndMean    | 193115140    | 6.024 ns/op        | 0 B/op                    |
|                                |             | CalculateSizeZScore       | 30467485     | 40.64 ns/op        | 0 B/op                    |
| `GOMAXPROCS=11 -benchtime=10s` | Apple M3 Pro | CalculateStdDevAndMean-11 | 1000000000   | 6.122 ns/op        | 0 B/op                    |
|                                |             | CalculateSizeZScore-11    | 299574828    | 40.35 ns/op        | 0 B/op                    |
| `GOMAXPROCS=11`| Apple M3 Pro | CalculateStdDevAndMean-11 | 192495468    | 6.024 ns/op        | 0 B/op                    |
|                                |             | CalculateSizeZScore-11    | 30869821     | 39.69 ns/op        | 0 B/op                    |