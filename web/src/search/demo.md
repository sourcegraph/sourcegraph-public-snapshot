# Great code search

Great code search helps you write better code more quickly. It helps you understand, debug, and reuse code. When you have great code search, you use it 5-10 times per day or more. It's right up there with your editor and code host.

At Sourcegraph, we think great code search is:

* Fast
* Up to date with all your code, so you never need to worry about indexing
* Powerful, with code intelligence and regular expression query support
* Relevant, with quick filtering and saved scopes
* Deeply integrated with your editor, your repository's history/blame info, and other tools

We built Sourcegraph's code search to be this great code search for all your organization's code. Let's give it a try.

Let's say I saw an error message "not found" and I'm trying to find where it comes from. I'll go to my organization's Sourcegraph and type it in surrounded by quotes to find exact matches. In under a second, it searched across hundreds of repositories and found a bunch of matches. I know it's coming from a code file, so I'll exclude Markdown files by adding `-file:.md`. And I know it's not in the website repository, so I'll add `-repo:website`. I'll narrow the search down further because I know the error message mentions "file" at the beginning, so I'll just add the term `file` to the beginning of the query. Now I've narrowed it to 3 results, and I can tell it's coming from this Java code. From here, I'll do a "find references" to see where this class is being used, and now I know where the error is coming from.

Now let's see how code search helps us find a specific function and real usage examples of it. I've got this function NewConn, so I'll just start by typing in `NewConn`. Looks like a lot of other similarly named functions are out there, so let me filter it down to just the definitions by searching for `func NewConn`. Now I'll exclude vendored code to just find original sources by selecting the "Non-vendor code" search scope. You can see that this is equivalent to adding these two `-file:` filters to the query. Now I see what I'm looking for! I can see references to it from its own repository, and also cross-repository references, like this one. I can even go to definition across repositories. And everything works instantly on any branch or commit, even from years ago.

Of course, this all needs to be integrated with your editor and other tools. And it is. Straight from my editor I can perform a search. And remember what we said about regular expressions? They're supported.

We want to build the best code search for your organization, and we think we've done it. Sourcegraph Server gives you code search that's fast, up-to-date, powerful, relevant, and deeply integrated. It's easy to set up and runs securely on your own network. Check out sourcegraph.com/server to try Sourcegraph Server for free.
