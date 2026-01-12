<p align="center">
    <img src="https://github.com/user-attachments/assets/ecf08358-3a49-4b16-95e3-20c4441cf9c0" width="45%" />
</p>
<p align="center">
    <img src="https://img.shields.io/badge/-1.23-brightgreen?style=plastic&logo=go&logoColor=white" />
	<img src="https://img.shields.io/badge/Svelte-FF3E00?style=plastic&logo=svelte&logoColor=white" />
</p>



## 🥀ToDo
- [ ] Optimize json updates
- [x] Collision mode ⛓️
- [x] Live mode
- [ ] Docker img (?)
- [x] Сorrect display of ANY types in the visualization
- [x] The ability to perform operations with the map directly from the visualization
- [ ] Implement map visualization for versions >1.23💀
- [ ] Add an endpoint with an overview of the work of the map (tutorial)
- [ ] I need a good image for the logo 🎃

## About Project
<img width="960" height="540" alt="{EB9CD144-2318-4688-922F-841DB8A6C077}" src="https://github.com/user-attachments/assets/6c94bb2e-5148-4a85-be9a-d71b6810e7f1" />

This is a simple program that visualizes the inner workings of the hashmap data structure in Golang using the unsafe package. We are talking about the old map (closed hashing) up to and including `version 1.23`. After version 1.23, the map began to work on a completely different principle. Version `1.23.12` was used in testing. You can read the original map code [here](https://cs.opensource.google/go/go/+/release-branch.go1.23:src/runtime/map.go).

## How to install&launch
```shell
git clone https://github.com/hyperocheck/golang-map-visualizer.git
cd golang-map-visualizer
go1.23 run ./cmd/visualizer/
```
## How to use
To set your hash map type, you just need to change the return type inside the anonymous function in the `main` func and, of course, return the map (file: `/cmd/visualizer/main.go`). Inside this anonymous function, you can do whatever you want with the map, for example, fill it with 100_000 elements or delete half of them. Here's what it looks like:
```go
func main() {

	// Create your map here and be sure to return it.
	// You can do anything with the map inside this block.
	// And also don't forget to specify the return type.
	fn := func(i_from, i_to int) map[int]float64 {  // <- TYPE

		m := make(map[int]float64)

		for i := i_from; i < i_to; i++ {
			m[i] = float64(i)
		}

		return m // <- RETURN MAP
	}
	// ------------------------------------

	work(fn)
}
```
Using the `--from` and `--to` flags, you can set the start and end of the iterator (if you want to use it at all).  
`go1.23 run ./cmd/visualizer --from 100 --to 1000`  
  
Actions in visualization:
Action |	Description
--- | ---
`double left click + d` |delete key
`double left click + (change value in side menu) + u` | update key

  
You have access to the cli with the following commands:

Command |	Description
--- | --- 
`show` | print map
`hmap` | print hmap structure
`exit` | exit from console & server down
`insert <k> <v>` | guess
`update <k> <v>` | guess
`delete <k>` | guess
## About custom K&V in map type (map[K]V)
**🥀PLEASE DO NOT FORGET TO MAKE ALL STRUCT FIELDS PUBLIC🥀**    
There are two ways to use custom types as keys and values. The __simple__ way is to just create a structure and add json tags to it. Then you will use the cli using the json input format. Something like this:
```go
// --- A SIMPLE WAY TO USE CUSTOM TYPES ---
// Simply create a custom structure with JSON tags. In the CLI, you will enter JSON. F.e.: (map[int]UserCustomDataExample1)
type UserCustomDataExample1 struct {
	I1 int    `json:"i1"`
	I2 int    `json:"i2"`
	S1 []bool `json:"s1`
}
// In the CLI, it will look like this: insert 100 {"i1":10, "i2":42, "s1": [true, false, false]} 
```
And __hard way__. You need to implement the Parse method for it so that you can use cli later. After all, the author can't guess how you want to enter your own structure in the insert command, right?:D For example, you want to use `int` as the key and such an interesting structure as the value 
```go
type MyInterestingStuct struct {
    I1 int
    I2 int
    S1 []int
}
```
How can we insert this value into the CLI? Let's see what you can come up with! For example, you came up with the following input format for this structure:  
`<int>,<int>;<bool, bool, ...>`  
All that remains is to implement the Parse method for this type, which will accept our string from the CLI as input and convert it into our structure for further serialization.
```go
func (UserCustomDataExample) Parse(s string) (UserCustomDataExample, error) {
	var result UserCustomDataExample

	s = strings.TrimSpace(s)

	// split ints ; bools
	parts := strings.SplitN(s, ";", 2)
	if len(parts) != 2 {
		return result, fmt.Errorf("expected format: <i1>,<i2>;<bool,bool,...>")
	}

	// --- parse ints ---
	intPart := strings.Split(parts[0], ",")
	if len(intPart) != 2 {
		return result, fmt.Errorf("expected two ints: <i1>,<i2>")
	}

	i1, err := strconv.Atoi(strings.TrimSpace(intPart[0]))
	if err != nil {
		return result, fmt.Errorf("invalid i1: %w", err)
	}

	i2, err := strconv.Atoi(strings.TrimSpace(intPart[1]))
	if err != nil {
		return result, fmt.Errorf("invalid i2: %w", err)
	}

	// --- parse bool slice ---
	boolPart := strings.TrimSpace(parts[1])
	boolStrs := []string{}
	if boolPart != "" {
		boolStrs = strings.Split(boolPart, ",")
	}

	s1 := make([]bool, len(boolStrs))
	for i, v := range boolStrs {
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return result, fmt.Errorf("invalid bool at index %d: %w", i, err)
		}
		s1[i] = b
	}

	result.I1 = i1
	result.I2 = i2
	result.S1 = s1

	return result, nil
}
```

## Easter eggs 😺
If you want to visualize a chain of blocks with a length of at least two, then use this formula to generate a certain number of items in a bucket: `(x * 8) * 0.8125`, x is the number of buckets (any number that is a power of two (min 8) -- 8, 16, 32, 64 ...). It will work on the 10th or 20th attempt, good luck:)
