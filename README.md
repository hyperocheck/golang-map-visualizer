
  <h2>Golang hashmap inspector</h2>
<p align="center">
	<img src="https://github.com/user-attachments/assets/6152ca31-f285-4cdc-ba73-d53078d85cd3" width="45%"/>
</p>

<p align="center">
	<img  src="https://img.shields.io/badge/-1.23-brightgreen?style=plastic&logo=go&logoColor=white">
</p>

## 🥀ToDo
- [ ] docker img
- [x] Сorrect display of ANY types in the visualization
- [ ] The ability to perform operations with the map directly from the visualization
- [ ] Implement map visualization for versions >1.23💀
- [ ] Add an endpoint with an overview of the work of the map (tutorial)

## About Project
<img width="1280" height="720" alt="image" src="https://github.com/user-attachments/assets/db7f49ba-f863-4e7c-b341-25d0c5a5be99" />

This is a simple program that visualizes the inner workings of the hashmap data structure in Golang using the unsafe package. We are talking about the old map (closed hashing) up to and including `version 1.23`. After version 1.23, the map began to work on a completely different principle. Version `1.23.12` was used in testing. You can read the original map code [here](https://cs.opensource.google/go/go/+/release-branch.go1.23:src/runtime/map.go).

## How to install&launch
```shell
git clone https://github.com/hyperocheck/golang-map-visualizer.git
cd golang-map-visualizer
go1.23 run ./cmd/visualizer/
```
## How to use
~~The only `GetUserMap` function is in the `/src/usermap/usermap.go` file. You need to create any map in it and return it. In this function, you can do whatever you want with the map before launching, for example, fill it in and remove half of the elements. By default, any custom key and value can be inserted into the map, but in order to use the console, you need to implement a parser and constructor of this type in the `/src/console/console.go` file. This means that you need to come up with an input format for your data type and write a parser for it yourself. It's best to use basic data types (like `int`, `string`, `bool` etc), since I haven't really bothered with outputting custom structures yet :)~~
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
