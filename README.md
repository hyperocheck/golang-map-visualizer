## About Project
<img width="960" height="540" alt="{1702FD80-222C-4F6E-91A3-AD43DB61C7CC}" src="https://github.com/user-attachments/assets/dcb2ec66-fb5d-4ebe-9f19-e503f082902d" />

This is a console application with web visualization that demonstrates the inner workings of the *map* data type in Go versions `<=1.23.x`. The original source code for Go maps can be found [here](https://cs.opensource.google/go/go/+/release-branch.go1.23:src/runtime/map.go). This project was developed and tested using version `1.23.12`, so using this specific version is highly recommended.

## How to download, install, and run
```shell
go install golang.org/dl/go1.23.12@latest
go1.23.12 download
git clone https://github.com/hyperocheck/golang-map-visualizer.git
cd golang-map-visualizer
go1.23.12 run ./cmd/visualizer/
```


## How to use
The entry point of the application is located in `./cmd/visualizer/main.go`. This is where a `map[string]int` is created, which you can pre-populate with data. You can use any serializable type as a key or value. Out of the box, the console fully supports `string`, `bool`, and `int` (including all its variations) in any key-value combination. If you want to use a custom data type while maintaining full console functionality, you must implement the following for your type:
1. The `FromIndex` method of the `Generatable` interface.
2. The `Parse` method of the `Parseable` interface.  

You can find an implementation example in `./cmd/visualizer/example.go`.
```go
// ./cmd/visualizer/main.go

func main() {    
    m := make(engine.Map[int, int])
    // m := make(engine.Map[string, int], 256)
    // m := make(engine.Map[string, CustomStruct], 1000)
    // m := make(engine.Map[CustomStuct, map[int]bool])

    work(m)
}
```
## How to play
Navigate to `http://localhost:8080` in your browser and open the CLI. The console supports command history (using ↑/↓ arrows) and Tab completion. Below are examples of available commands and their descriptions.
```shell
# Insert a key-value pair
# Equivalent to: map[123] = 382783728
insert 123 382783728
```
```shell
# Update a key-value pair
# Equivalent to: map[123] = 9990
update 123 9990
```
```shell
# Delete a key-value pair
# Equivalent to: delete(map, 123)
delete 123
```
```shell
# Display all key-value pairs
show
```
```shell
# Display the current state of the hmap struct
hmap
```
```shell
# Show available commands
help
```
```shell
# Exit the program
# You can also use Ctrl+C
exit
```
```shell
# Insert key-value pairs in a loop from 10 to 1400
# Equivalent to: for i := 10; i <= 1400; i++ { map[i] = i }
range insert 10 1400
```
```shell
# Insert key-value pairs from 10 to 1400 using live mode
# Equivalent to: for i := 10; i <= 1400; i++ { map[i] = i }
range insert 10 1400 --life
```
```shell
# Delete key-value pairs in a loop from 10 to 1400
# Equivalent to: for i := 10; i <= 1400; i++ { delete(map, i) }
range delete 10 1400
```
```shell
# Collision mode
# Specifically insert 10 key-value pairs into the first bucket (bid=0)
evil 10
```
```shell
# Collision mode
# Specifically insert 10 key-value pairs into the first bucket (bid=0) in live mode
evil 10 --life
```
```shell
# Collision mode
# Specifically insert 10 key-value pairs into the 3rd bucket (bid=3) in live mode
evil 10 --bid 3 --life
```
```shell
# Show a detailed step-by-step process of searching for a value by key
# Matches the actual logic of the original mapaccess1 function
mapaccess1 2382
```



