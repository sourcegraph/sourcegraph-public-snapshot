# Code Insights repository scope

Prior to Sourcegraph 4.5, on insight creation you have the option to:
* Specify a list of repositories to run your code insight over
* Run your insight over all repositories

From 4.5 the "Run your insight over all repositories" checkbox has been replaced with a search query box.

You can use repository filters as you would in a normal Sourcegraph query to specify which repositories you want to run your code insight over.
The creation form will display how many repositories your search query matches and you can preview the matches. 

If your query matches less than 20 repos a live preview will also be displayed.

## How is the list of repositories resolved?



