/*
Package mafsa implements Minimal Acyclic Finite State Automata (MA-FSA)
in a space-optimized way as described by Dacuik, Mihov, Watson, and
Watson in their paper, "Incremental Construction of Minimal Acyclic
Finite-State Automata" (2000). It also implements Minimal Perfect Hashing (MPH)
as described by Lucceshi and Kowaltowski in their paper, "Applications of
Finite Automata Representing Large Vocabularies" (1992).

Unscientifically speaking, this package lets you store large amounts of
strings (with Unicode) in memory so that membership queries, prefix lookups,
and fuzzy searches are fast. And because minimal perfect hashing is included,
you can associate each entry in the tree with more data used by your application.
See the README or the end of this documentation for a brief tutorial.

MA-FSA structures are a specific type of Deterministic Acyclic Finite
State Automaton (DAFSA) which fold equivalent state transitions into each
other starting from the suffix of each entry. Typical construction algorithms
involve building out the entire tree first, then minimizing the completed
tree. However, the method described in the paper above allows the tree
to be minimized after every word insertion, provided the insertions are
performed in lexicographical order, which drastically reduces memory usage
compared to regular prefix trees ("tries").

The goal of this package is to provide a simple, useful, and correct
implementation of MA-FSA. Though more complex algorithms exist for removal
of items and unordered insertion, these features may be outside the scope
of this package.

Membership queries are on the order of O(n), where n is the length of the
input string, so basically O(1). It is advisable to keep n small since
long entries without much in common, especially in the beginning or end of the
string, will quickly overrun the optimizations that are available. In those
cases, n-gram implementations might be preferable, though these will use more
CPU.

This package provides two kinds of MA-FSA implementations. One, the BuildTree,
facilitates the construction of an optimized tree and allows ordered insertions.
The other, MinTree, is effectively read-only but uses significantly less memory
and is ideal for production environments where only reads will be occurring.

Usually your build process will be separate from your production application,
which will make heavy use of reading the structure.

To use this package, create a BuildTree and insert your items in lexicographical
order:

    bt := mafsa.New()
    bt.Insert("cities") // an error will be returned if input is < last input
    bt.Insert("city")
    bt.Insert("pities")
    bt.Insert("pity")
    bt.Finish()

The tree is now compressed to a minimum number of nodes and is ready to be
saved.

    err := bt.Save("filename")
    if err != nil {
    	log.Fatal("Could not save data to file:", err)
    }

In your production application, then, you can read the file into a MinTree
directly:

    mt, err := mafsa.Load("filename")
    if err != nil {
    	log.Fatal("Could not load data from file:", err)
    }

The mt variable is a *MinTree which has the same data as the original BuildTree,
but without all the extra "scaffolding" that was required for adding new elements.

The package provides some basic read mechanisms.

    // See if a string is a member of the set
    fmt.Println("Does tree contain 'cities'?", mt.Contains("cities"))
    fmt.Println("Does tree contain 'pitiful'?", mt.Contains("pitiful"))

    // You can traverse down to a certain node, if it exists
    fmt.Printf("'y' node is at: %p\n", mt.Traverse([]rune("city")))

    // To traverse the tree and get the number of elements inserted
    // before the prefix specified
    _, idx := mt.IndexedTraverse([]rune("pit"))
    fmt.Println("Elements inserted before entries starting with 'pit':", idx)
*/
package mafsa
