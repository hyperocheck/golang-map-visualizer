
<h2>Golang hashmap inspector</h2>
<p align="center">
	<img src="https://github.com/user-attachments/assets/b9ad1609-bd4e-4647-8647-f84412933365" width="45%/>
</p>

<p align="center">
	<img src="https://img.shields.io/badge/-1.23-brightgreen?style=plastic&logo=go&logoColor=white">
</p>

## 🥀ToDo
- [ ] docker img
- [x] Сorrect display of ANY types in the visualization
- [x] The ability to perform operations with the map directly from the visualization
- [ ] Implement map visualization for versions >1.23💀
- [ ] Add an endpoint with an overview of the work of the map (tutorial)

## About Project
<img width="960" height="540" alt="{A61A0BC2-D31B-487D-9580-CFB0C3DE84AA}" src="https://github.com/user-attachments/assets/8441d3ff-c574-4369-8983-1199dd4fbbd8" />

This is a simple program that visualizes the inner workings of the hashmap data structure in Golang using the unsafe package. We are talking about the old map (closed hashing) up to and including `version 1.23`. After version 1.23, the map began to work on a completely different principle. Version `1.23.12` was used in testing. You can read the original map code [here](https://cs.opensource.google/go/go/+/release-branch.go1.23:src/runtime/map.go).

## How to install&launch
```shell
git clone https://github.com/hyperocheck/golang-map-visualizer.git
cd golang-map-visualizer
go1.23 run ./cmd/visualizer/
```
## How to use
To set the type of your hashmap, you need to return it from the function call `engine.Start` in main func (`/cmd/visualizer/main.go`). It's all. Inside this function, you can pre-do whatever you want with it. You can specify as a key and value everything that allows Golang and everything that can be serialized with the built-in `encoding/json` package:). If a standard package can serialize it, then there will be no problems with visualization. Here is an example where I set a custom structure as a value. 
```go
type MyCustomData struct {
	Label bool
	Map   map[int][]string
}

func main() {
	preview.Preview()

	// -------------- You're editing here. --------------
	usermapo := engine.Start(func(iters int, maxChain bool) map[int]MyCustomData { // <- DON'T FORGET TO SPECIFY THE TYPE OF THE RETURNED MAP.

		m := make(map[int]MyCustomData)

		for i := range 100 {
			m[i] = GenerateMyCustomData()
		}

		return m // <- DON'T FORGET TO RETURN THE CARD.
	})
	// ---------------------------------------------------

    // ... my code
}
```

You have access to the cli with the following commands:

Command |	Description
--- | --- 
`show` | print map
`hmap` | print hmap structure
`exit` | exit from console & server down
`insert <k> <v>` | guess
`update <k> <v>` | guess
`delete <k>` | guess
## Easter eggs 😺
If you want to visualize a chain of blocks with a length of at least two, then use this formula to generate a certain number of items in a bucket: `(x * 8) * 0.8125`, x is the number of buckets (any number that is a power of two (min 8) -- 8, 16, 32, 64 ...). It will work on the 10th or 20th attempt, good luck:)
