# Introduction to Code Insights 

## Overview

TODO TODO image

Code insights is the first feature in Sourcegraph that can tell you things about your code at a high level.

Code Insights can answer questions like “How is a migration progressing?”, “What areas of the code are most vulnerable to bugs?”, and “How many developers are using a specific API?” Code Insights will (later in 2021) also incorporate third-party data like code coverage or static analysis metrics to deliver on the promise of aggregating everything you can know about your code.

Sourcegraph is in the unique position to provide these insights because we have universal code search: to know anything about your code at a high level accurately means you must know everything about your code at a low level. These include: 

- How your code is tracking against any migration, pattern, or code smell goals
- How your code is changing over time and what areas may need more or less developer attention
- Understanding your code’s current and historical content, like its languages, libraries, and structure
- What patterns or outliers exist in your third party tools’ data when viewed at a high level
- Any of the above conceptss, but also filtered by repository, engineering team, or other division

## Concepts (Terminology)

- A **Code Insight** is a single individual chart, graph, or other quantitative display 
- TODO TODO explain the types here

## Future Plans

Code Insights currently answers search-based questions and language usage questions about your codebase. 

In late Spring 2021, Code Insights will have additional display location and viewer-management settings. 

In Summer 2021, Code Insights will scale to answer these same questions for all of your repositories or your entire codebase in the same insight. 

In Fall 2021, Code Insights will have prototype versions of 3rd-party data integration that let you sync your other tools to display in code insights. 

In late 2021, Code Insights will provide additional filtering and analysis controls. 

## Known Issues 

- In Sourcegraph 3.28 or earlier: code insights can only scale to a subset of all your repositories. Trying to run a search-based insight (LINK TODO TODO) on greater than ~20 repositories will time out. 
