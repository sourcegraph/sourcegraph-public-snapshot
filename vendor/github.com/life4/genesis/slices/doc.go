// 🍞 Package slices provides generic functions for slices.
//
// The package is inspired by [Enum] and [List] Elixir modules.
//
// # Conventions
//
//   - All functions accepting a slice accept it as the very first argument.
//   - If a function provides 2 implementations one of which accepts a function
//     (for example, [Equal] and [EqualBy]), the one accepting the function
//     has suffix "By".
//   - If a function is concurrent, it has suffix "Async".
//   - Almost all functions are pure and don't modify the given slice.
//     The only exception so far is [Shuffle].
//
// # Functions
//
// This package has a lot of functions and it might be hard to find what you need
// if you don't know how it is called. Because of that, this section provides
// an easier to navigate list of all functions defined here. Each of them is grouped
// based on the return type.
//
// Also, for easier visual navigation, a signature is provided for each function
// where instead of argument type and name an emoji is used:
//
//   - 📚 is a slice
//   - 📕 is a slice element
//   - 💬 is a function
//   - ❓ is a bool
//   - 📺 is a channel
//   - 🗺 is a map
//   - 💥 is an error
//   - 🔢 is an int
//   - 🎲 is a randomization seed
//   - 🧑‍🔧️ is a number of workers
//
// 🎲 Randomization functions:
//
//   - [Choice](📚, 🎲) (📕, 💥)
//   - [Shuffle](📚, 🎲)
//   - [TakeRandom](📚, 🔢, 🎲) (📚, 💥)
//
// ❓ Functions returning a bool:
//
//   - [All](📚, 💬) ❓
//   - [AllAsync](📚, 🧑‍🔧️, 💬) ❓
//   - [Any](📚, 💬) ❓
//   - [AnyAsync](📚, 🧑‍🔧️, 💬) ❓
//   - [Contains](📚, 📕) ❓
//   - [EndsWith](📚, 📕) ❓
//   - [Equal](📚, 📚) ❓
//   - [EqualBy](📚, 📚, 💬) ❓
//   - [Sorted](📚) ❓
//   - [SortedUnique](📚) ❓
//   - [Same](📚) ❓
//   - [StartsWith](📚, 📕) ❓
//   - [Unique](📚) ❓
//
// 🗺 Functions returning a map:
//
//   - [GroupBy](📚, 💬) 🗺
//   - [ToKeys](📚, 📕) 🗺
//   - [ToMap](📚) 🗺
//
// 📺 Functions returning a channel:
//
//   - [Cycle](📚) 📺
//   - [Permutations](📚, 🔢) 📺
//   - [Product](📚, 🔢) 📺
//   - [Product2](...📚) 📺
//   - [ToChannel](📚) 📺
//   - [Zip](...📚) 📺
//
// 📕 Functions returning a single item:
//
//   - [Find](📚, 💬) (📕, 💥)
//   - [Last](📚) (📕, 💥)
//   - [Max](📚) (📕, 💥)
//   - [Min](📚) (📕, 💥)
//   - [Reduce](📚, 📕, 💬) 📕
//   - [ReduceAsync](📚, 🧑‍🔧️, 💬) 📕
//   - [ReduceWhile](📚, 📕, 💬) (📕, 💥)
//   - [Sum](📚) 📕
//
// 🔢 Functions returning an int:
//
//   - [Count](📚, 📕) 🔢
//   - [CountBy](📚, 💬) 🔢
//   - [FindIndex](📚, 💬) 🔢
//   - [Index](📚, 📕) (🔢, 💥)
//   - [IndexBy](📚, 💬) (🔢, 💥)
//
// 🖨 Functions that take a slice and return a slice:
//
//   - [Copy](📚) 📚
//   - [Dedup](📚) 📚
//   - [DropZero](📚) 📚
//   - [Reverse](📚) 📚
//   - [Shrink](📚) 📚
//   - [Sort](📚) 📚
//   - [Uniq](📚) 📚
//
// 📚 Functions returning a new slice:
//
//   - [ChunkBy](📚, 💬) 📚
//   - [ChunkEvery](📚, 🔢) (📚, 💥)
//   - [Concat](...📚) 📚
//   - [DedupBy](📚, 💬) 📚
//   - [Delete](📚, 📕) 📚
//   - [DeleteAll](📚, 📕) 📚
//   - [DeleteAt](📚, 🔢) (📚, 💥)
//   - [Difference](📚, 📚) 📚
//   - [DropEvery](📚, 🔢, 🔢) (📚, 💥)
//   - [DropWhile](📚, 💬) 📚
//   - [Filter](📚, 💬) 📚
//   - [FilterAsync](📚, 🧑‍🔧️, 💬) 📚
//   - [Grow](📚, 🔢) 📚
//   - [InsertAt](📚, 🔢, 📕) (📚, 💥)
//   - [Intersect](📚, 📚) 📚
//   - [Intersperse](📚, 📕) 📚
//   - [Map](📚, 💬) 📚
//   - [MapAsync](📚, 🧑‍🔧️, 💬) 📚
//   - [MapFilter](📚, 💬) 📚
//   - [Prepend](📚, ...📕) 📚
//   - [Reject](📚, 💬) 📚
//   - [Repeat](📚, 🔢) 📚
//   - [Replace](📚, 🔢, 🔢, 📕) (📚, 💥)
//   - [Scan](📚, 📕, 💬) 📚
//   - [SortBy](📚, 💬) 📚
//   - [Split](📚, 📕) 📚
//   - [TakeEvery](📚, 🔢, 🔢) (📚, 💥)
//   - [TakeWhile](📚, 💬) 📚
//   - [Union](📚, 📚) 📚
//   - [Window](📚, 🔢) (📚, 💥)
//   - [Without](📚, 📕) 📚
//   - [Wrap](📕) 📚
//
// 😶 Functions returning a something else or nothing:
//
//   - [Each](📚, 💬)
//   - [EachAsync](📚, 🧑‍🔧️, 💬)
//   - [EachErr](📚, 💬) 💥
//   - [Join](📚, string) string
//   - [Partition](📚, 💬) (📚, 📚)
//
// [Enum]: https://hexdocs.pm/elixir/1.12/Enum.html
// [List]: https://hexdocs.pm/elixir/1.12/List.html
package slices
