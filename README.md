<p align="center">
    <img src="https://github.com/user-attachments/assets/ecf08358-3a49-4b16-95e3-20c4441cf9c0" width="45%" />
</p>
<p align="center">
    <img src="https://img.shields.io/badge/-1.23-brightgreen?style=plastic&logo=go&logoColor=white" />
</p>

## Todo
- [ ] Implement map visualization for swisstables (in progress, branch swisstables)
- [ ] Add an endpoint with an overview of the work of the map (or .md file)
- [ ] I need a good image for the logo

## About Project
<img width="888" height="508" alt="{5847B7A3-78FE-4037-B16B-C85C607C0CC4}" src="https://github.com/user-attachments/assets/85bc1411-3ab7-4797-bd67-e7ef1a25fb48" />

This is a simple program that visualizes the inner workings of the hashmap data structure in Golang using the unsafe package. We are talking about the old map (closed hashing) up to and including `version 1.23`. After version 1.23, the map began to work on a completely different principle. Version `1.23.12` was used in testing. You can read the original map code [here](https://cs.opensource.google/go/go/+/release-branch.go1.23:src/runtime/map.go).

## How to install&launch
```shell
git clone https://github.com/hyperocheck/golang-map-visualizer.git
cd golang-map-visualizer
go1.23 run ./cmd/visualizer/
```
## How to use
`./src/visualizer/main.go`
```go
func main() {
    // так создается мапа, тут же указывается тип
    // тут можно заранее сделать чо угодно с мапой
    // также есть флаги --from и --to, можете использовать их при запуске
    // типо go1.23 run ./cmd/visualizer --from 100 --to 1000 
	m := make(engine.Map[int, int], 1)

	for i := cmd.Flag.From; i < cmd.Flag.To; i++ {
		m[i] = i
	}

	work(m) // не удалять:))
}

```
## How to play


