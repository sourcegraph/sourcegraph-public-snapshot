MA-FSA for Go
=============

Package mafsa implements Minimal Acyclic Finite State Automata (MA-FSA) with Minimal Perfect Hashing (MPH). Basically, it's a set of strings that lets you test for membership, do spelling correction (fuzzy matching) and autocomplete, but with higher memory efficiency than a regular trie. With MPH, you can associate each entry in the tree with data from your own application.

In this package, MA-FSA is implemented by two types:

- BuildTree
- MinTree

A BuildTree is used to build data from scratch. Once all the elements have been inserted, the BuildTree can be serialized into a byte slice or written to a file directly. It can then be decoded into a MinTree, which uses significantly less memory. MinTrees are read-only, but this greatly improves space efficiency.


## Tutorial

Create a BuildTree and insert your items in lexicographical order. Be sure to call `Finish()` when you're done.

```go
bt := mafsa.New()
bt.Insert("cities") // an error will be returned if input is < last input
bt.Insert("city")
bt.Insert("pities")
bt.Insert("pity")
bt.Finish()
```

The tree is now compressed to a minimum number of nodes and is ready to be saved.

```go
err := bt.Save("filename")
if err != nil {
    log.Fatal("Could not save data to file:", err)
}
```

In your production application, then, you can read the file into a MinTree directly:

```go
mt, err := mafsa.Load("filename")
if err != nil {
    log.Fatal("Could not load data from file:", err)
}
```

The `mt` variable is a `*MinTree` which has the same data as the original BuildTree, but without all the extra "scaffolding" that was required for adding new elements. In our testing, we were able to store over 8 million phrases (average length 24, much longer than words in a typical dictionary) in as little as 2 GB on a 64-bit system.

The package provides some basic read mechanisms.

```go
// See if a string is a member of the set
fmt.Println("Does tree contain 'cities'?", mt.Contains("cities"))
fmt.Println("Does tree contain 'pitiful'?", mt.Contains("pitiful"))

// You can traverse down to a certain node, if it exists
fmt.Printf("'y' node is at: %p\n", mt.Traverse([]rune("city")))

// To traverse the tree and get the number of elements inserted
// before the prefix specified
node, idx := mt.IndexedTraverse([]rune("pit"))
fmt.Println("Index number for 'pit': %d", idx)
```

To associate entries in the tree with data from your own application, create a slice with the data in the same order as the elements were inserted into the tree:

```go
myData := []string{
    "The plural of city",
    "Noun; a large town",
    "The state of pitying",
    "A feeling of sorrow and compassion",
}
```

The index number returned with `IndexedTraverse()` (usually minus 1) will get you to the element in your slice if you traverse directly to a final node:

```go
node, i := mt.IndexedTraverse([]rune("pities"))
if node != nil && node.Final {
    fmt.Println(myData[i-1])
}
```

If you do `IndexedTraverse()` directly to a word in the tree, you must -1 because that method returns the number of elements in the tree before those that *start* with the specified prefix, which is non-inclusive with the node the method landed on.

There are many ways to apply MA-FSA with minimal perfect hashing, so the package only provides the basic utilities. Along with `Traverse()` and `IndexedTraverse()`, the edges of each node are exported so you may conduct your own traversals according to your needs.


## Encoding Format

This section describes the file format used by `Encoder` and `Decoder`. You probably will never need to implement this yourself; it's already done for you in this package.

BuildTrees can be encoded as bytes and then stored on disk or decoded into a MinTree. The encoding of a BuildTree is a binary file that is composed of a sequence of words, which is a sequence of big-endian bytes. Each word is the same length. The file is inspected one word at a time.

The first word contains file format information. Byte 0 is the file version (right now, 1). Byte 1 is the word length. Byte 2 is the character length. Byte 3 is the pointer length. The rest of the bytes of the first word are 0s.

The word length must be at least 4, and must equal Byte 2 + Byte 3 + 1 (for the flag byte).

Package smartystreets/mafsa initializes the file with this word:

    []byte{0x01 0x06 0x01 0x04 0x00 0x00}

This indicates that using version 1 of the file format, each word is 6 bytes long, each character is 1 byte, and each pointer to another node is 4 bytes. This pointer size allows us to encode trees with up to 2^32 (a little over 4.2 billion) edges.

Every other word in the file represents an edge (not a node). Those words consist of bytes according to this format:

    [Character] [Flags] [Pointer]

The length of the character and pointer bytes are specified in the initial word, and the flags part is always a single byte. This allows us to have 8 flags per edge, but currently only 2 are used.

Flags are:

- `0x01` = End of Word (EOW, or final)
- `0x02` = End of Node (EON)

A node essentially consists of a consecutive, variable-length sequence of words, where each word represents an edge. To encode a node, write each edge sequentially. Set the final (EOW) flag if the node it points to is a final node, and set the EON flag if it is the last edge for that node. The EON flag indicates the next word is an edge belonging to a different node. The pointer in each word should be the *word* index of the start of the node being pointed to.

For example, the tree containing these words:

- cities
- city
- pities
- pity

Encodes to:

    0   01 06 0104 0000
    1   63 00 0000 0003
    2   70 02 0000 0004
    3   69 02 0000 0005
    4   69 02 0000 0005
    5   74 02 0000 0006
    6   69 00 0000 0008
    7   79 03 0000 0000
    8   65 02 0000 0009
    9   73 03 0000 0000

Or here's the hexdump:

    00000000  01 06 01 04 00 00 63 00  00 00 00 03 70 02 00 00  |......c.....p...|
    00000010  00 04 69 02 00 00 00 05  69 02 00 00 00 05 74 02  |..i.....i.....t.|
    00000020  00 00 00 06 69 00 00 00  00 08 79 03 00 00 00 00  |....i.....y.....|
    00000030  65 02 00 00 00 09 73 03  00 00 00 00              |e.....s.....|

Now, this tree isn't a thorough-enough example to test your implementation against, but it should illustrate the idea. First notice that the first word is the initializer word. The second word is the first edge coming off the root node, and words 1 and 2 are both edges coming off the root node. The word at 2 has the EON flag set (0x02) which indicates the end of the edges coming off that node. You'll see that edges at word indices 3 and 4 both point to the node starting at edge with word index 5. That would be the shared 't' node after 'ci' and 'pi'.

If a node has no outgoing edges, the pointer bits are 0.