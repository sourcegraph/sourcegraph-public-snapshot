Issue link:

https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=accessibility%2Cwcag%2F2.1%2Cwcag%2F2.1%2Fauditing%2C&title=%5BAccessibility%5D%3A+

# 1

[Accessibility Audit] Search: Making a search

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

Submit search:

- I can type a query in the search box and submit the search
- I can type a query in the search box, I can get suggestions for search filters, repositories, and other common search items. I can select these suggestions with either the keyboard or the mouse.
- I can distinguish between different types of search terms in the search box, including filters and boolean operators.

Additional search controls:

- I can use the toggles next to the search box to enable or disable case sensitivity, regex pattern matching, or structural search in my search.
- I can use the Copy Query button to copy the full contents of my query to the clipboard, including hidden parameters represented by the toggles and search context dropdown..

Utilizing search contexts:

- I can open the Search Contexts dropdown to view a list of all the contexts I have access to and click on one to change what search context I am searching through; I can open Sourcegraph later and my selected search context is preserved; I can click the "Reset" button on the dropdown to revert back to the default context.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 2

[Accessibility Audit] Search: (Authenticated) Search homepage

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

- As an authenticated user, I can load the search homepage and see my recent activity, including recently searched repositories, recent search queries, and recently viewed files. I can click on any of these to execute the search or open the file.
- As an authenticated user on Sourcegraph Server, I can open the search homepage and see saved searches. I can toggle between viewing only my saved searches and all saved searches I have access to. I can click on any saved search to execute it.
- As an authenticated user on Sourcegraph Server, I can open the search homepage and see all the community search contexts available on Sourcegraph Cloud. I can click on any of them to navigate to a tailored search homepage for that community.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 3

[Accessibility Audit] Search: Search results

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

After executing a search:

- I can see search results as they come in; I don't have to wait for the search to complete to start seeing results.
- I can view search progress, the number of results, and any errors, warnings or information on skipped results.
- I can click a button to start the creation of a new search context, code monitor, or saved search based on my query.
- For all results, I can see the result type (file, path, symbo, repository, diff, or commit message), repository name, code host, stars, and the last time that repository was synced with Sourcegraph. I can click on the repo name to go to that repo's page.
- For file, path, and symbol results, I can see the file's path. I can click on the file path to go to that file's page.
- For file results, I can see the number of content matches and a snippet of the content matches. If too many content matches exist for the file, only a portion of them will be shown; I can click a button to show all of them and revert back; I can also click a button to expand/contract all snippets for all results. I can click on any snippet to go to the file's page on the snippet locations. I can select text in the snippet and copy it to the clipboard with either the context menu or the keyboard.
- For symbol results, I can see the number of matches for the file and the list of matches with the symbol's type and name. I can click on a symbol to go to the file's page on the symbol's location.
- For repository results, I can see the repository's description from its code host.
- For diff and commit message results, I can see the author, commit title, commit short hash, and relative timestamp and I can click on any of these to go to the commit's page. I can see snippets that match the search and I can click on these to go to the commit's page.
- For diff results, I can see in the snippets which file they were found in and if it is an add or remove change.
- I can use the search sidebar to modify my search:
  - To only find specific types of results; for example, repositories or commit messages.
  - To add filters suggested based on the results of my search; for example, filter by one of the different programming languages in the results.
  - To add a repository filter based on the results of my search.
  - To add any filter or search operator; if I don't know what each filter does, I can easily learn it by viewing a short description.
  - To add a custom search snippet I have defined in my settings.
- If I have no results or have an error in my search, I can get helps and hints on how to modify my search to fix any issues and help find what I’m looking for.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 4

[Accessibility Audit] Search: Search Contexts Management

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

- On any search page, I can open the search contexts dropdown and click on "Manage context” to open the search contexts management page.
- Anywhere in Sourcegraph, I can use the “Code Search” > “Contexts” menu item to open the search contexts management page.
- On the search contexts management page, I can view a list of all the contexts I have access to; I can sort and filter the list to help find what I need if the list is very long. For each context, I can see its name, description, query (if it is query-based), number of repositories (if it is repository based), and when it was last updated. I can click on the context’s name to open the context’s page.
- I can create a new search context by clicking on the “Create search context” button. I can select myself, one of my org, or (if I'm an admin) the global scope as the owner of the context. I can give the context a name and a description, and select whether the context should be public or private. I can use either a special search query or a JSON with a list of repositories and revisions to create the context; for either of these, I get autocomplete suggestions, syntax highlighting, and validation.
- I can edit a context by clicking on its name and then clicking on “Edit”. I can change anything about the context in the edit view, except the owner. I can also delete the context.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 5

[Accessibility Audit] Search: Code Monitoring - View

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

General information:

- From any page in Sourcegraph, I can use the “Monitoring” menu item to open the code monitoring management page. If I don’t have any code monitors, I am redirected to the getting started page.
- I can use the tabs in the code monitoring page to switch between the code monitor list, the getting started view, and the logs list.
- I can use the “Create code monitor” button to go to the create monitor page.

Listing code monitors:

- On the code monitor list, I can see a list of code monitors with some details, including the name of the monitor, what actions it will take when the trigger occurs, and whether it is enabled.
- On the code monitor list, I can use the filters to see all code monitors I have access to.
- On the code monitor list, I can use the filters to see all code monitors owned by me.
- On the code monitor list, I can click on a code monitor in order to edit it. This takes me to the edit code monitor flow.
- On the code monitor list, I can disable a code monitor without the need to open the edit page.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 6

[Accessibility Audit] Search: Code Monitoring - Getting started

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

- On the getting started page, I can use the button that says “Create code monitor” to create a new code monitor. This goes to the create code monitor flow.
- On the getting started page, I can click on “create copy of monitor” on the example code monitors cards in order to create a copy of the code monitor in my code monitors. There are a variety of examples that I can use.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 7

[Accessibility Audit] Search: Code Monitoring - Logs

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

- On the logs page, I can click on the link that says “Monitor details” for any code monitor and see the edit page for that code monitor.
- On the logs page, I can click on a row in the table to see more details about that code monitor’s runs, including when it was run and if the run was successful or had any errors.
- Within the entry for an individual code monitor, I can also click to see more details about each run, including which triggers and actions were run and if there were any errors. I can also click on the link with the number of results to see the search associated with the code monitor on Sourcegraph.
- Within the entry for a particular trigger or action, I can click on it to see more details about the monitor trigger and the monitor action. Example details include error messages.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 8

[Accessibility Audit] Search: Code Monitoring - Create/Edit

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

- On the create and edit code monitor page, I can navigate back to the code monitors landing page by using the code monitor icon
- On the create and edit code monitor page, I can go to documentation about code monitors by clicking on the link that says “learn more.”
- On the create code monitor page, I can add a name for the code monitor in the input field.
- I cannot change the owner of the code monitor; this is by design.
- On the create and edit code monitor page, I can add a search result trigger to the code monitor using the card.
  - This entails adding a search query in the search query input. The query input will give autocomplete suggestions and syntax highlighting. I can use the “test query” button to open a search results page with this query to verify it works. The search query has certain limitations that must be fulfilled and I can easily know if my query is valid under these limitations.
  - I can confirm this query with the “continue” button
  - I can cancel my changes with the “cancel” button
- If a trigger has been added, I can add actions to the code monitor using the cards.
  - The actions are “send email notifications,” “send Slack message to channel,” and “call a webhook”
  - Each action type opens a new card
- Send email notifications
  - The “recipients” field cannot be changed; this is by design
  - I can send a test email to the email address in the recipients field by using the button labeled “Send test email”
  - I can change the state of the action by using the toggle; the label will reflect the state of the toggle
  - I can use the button labeled “continue” to progress with creating the code monitor
  - I can use the button labeled “cancel” to not add the action
  - I can use the button labeled “delete” to remove the action
- Send Slack message to channel
  - I can click on the link labeled “Slack” to go to the Slack apps website
  - I can go to instructions about setting up Slack messages by clicking on the link labeled “Read more about how to set up Slack webhooks in the docs”
  - I can add a webhook URL in the input labeled “Webhook URL”
  - If a webhook URL has been added, I can send a test message to the Slack channel identified by the webhook URL by using the button labeled “Send test message”
  - I can change the state of the action by using the toggle; the label will reflect the state of the toggle
  - I can use the button labeled “continue” to progress with creating the code monitor
  - I can use the button labeled “cancel” to not add the action
  - I can use the button labeled “delete” to remove the action
- Call a webhook
  - I can add a webhook URL in the input labeled “Webhook URL”
  - If a webhook URL has been added, I can send a test message target identified by the webhook URL by using the button labeled “Call webhook with test payload”
  - I can change the state of the action by using the toggle; the label will reflect the state of the toggle
  - I can use the button labeled “continue” to progress with creating the code monitor
  - I can use the button labeled “cancel” to not add the action
  - I can use the button labeled “delete” to remove the action
- The code monitor can be toggled on or off before saving
  - The default state is “active”
- Once a name, trigger, and action have been added, the code monitor can be saved
  - “Cancel” will not create the code monitor or undo the edits and return to the landing page
- Button labeled “Delete” will delete the code monitor

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 9

[Accessibility Audit] Search: Notebooks

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

- On the \*/notebooks page, I can:
  - Create a new Notebook
  - Import a Markdown file to create a new notebook
  - Search search the title and Markdown block content of Notebooks I created
  - Use tabs to toggle between displaying notebooks based on which ones I own, which ones I’ve starred, which ones my organization owns, and all the notebooks on my instances (Explore tab).
  - For every notebook on the list, I can see its name, description, owner, number of blocks, number of stars, and creation and update dates. I can click on a notebook to open it.
  - When creating or opening a notebook I have edit permissions to:
    - I can edit the title of the Notebook
    - I can delete the Notebook from the action menu
    - I can change the sharing permissions
    - I can add one or more of the following block types via the icon buttons, typing any string and selecting an available option from the dropdown that appears, or typing “/” and then type of block they want to create:
      - Markdown
        - I can use standard Markdown formatting such as lists, headings, and code formatting,
      - Sourcegraph search query
        - I can use the full Sourcegraph query language to generate search results as if I was in the main Sourcegraph app
      - File or code snippet
        - I can use the Sourcegraph query language to find a file of interest and optionally select a line range of interest
        - I can paste in a URL with a range of one or more lines
      - Symbol
        - I can use the Sourcegraph query language to search for symbols, optionally using the repo and file path filters to narrow my search
    - I can make a copy of the Notebook to My Notebooks
    - I can export the Notebook as a Markdown file
    - I can favorite the Notebook
    - I can run all runnable blocks in the Notebook by clicking the “Run all blocks” button
    - I can move blocks up or down in the list of blocks by clicking the “move up” or “move down” buttons (or Cmd/Ctrl + up arrow or Cmd/Ctrl + down arrow)
    - I can delete blocks by clicking the “Delete” button or Cmd/Ctrl + backspace
    - I can duplicate a block by clicking the “Duplicate” button or Cmd/Ctrl + D
    - I can save and render a block with Cmd/Ctrl + Return
    - I can edit a block by click the “Edit” button or pressing Return when a block is selected
    - I can open the full contents of any block results in a new tab by clicking the “Open in new tab” button
    - I can change the sharing settings of the Notebook to:
      - Share it with an organization of which I’m a member, which allows any member of the organization to edit it
      - Share it with an organization of which I’m a member and make it viewable to anyone on the instance, but editable only by members of my organization
      - Share it with my entire instance, but only I can edit
      - Only I can view or edit the notebook
  - When opening a notebook I only have read-access to:
    - I can view all blocks in the notebook and run searches
    - I can export and copy the notebook
    - I cannot change any of the block contents, block order, or metadata about the notebook

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 10

[Accessibility Audit] Search: Notepad

### Use cases to replicate

Taken from the [user journey list document](https://docs.google.com/document/d/1IA90OFew5MTcxFndKS2Xn4vb8sOt68sj537P7AGj_Tk/edit?usp=sharing). Use this for further context.

Please take these following use cases and replicate them when running your audit.

Perquisite: The notepad is enabled in settings

- When I run a search:
  - I see the notepad component in the bottom right corner of the screen
  - I can add a search to the notepad by clicking the “+ search” button in the notepad
    - When I add a search I can optionally add an annotation related to it
  - When I view a file:
    - I can add a file to the notepad by clicking the “+ file” button in the notepad
      - When I add a file I can optionally add an annotation related to it
    - If one or more lines are selected, I can add a file range to the notepad by clicking the “+ range” button in the notepad
      - When I add a range I can optionally add an annotation to it
  - I can delete an item added to the notepad by clicking the “x” icon in the item component
  - I can add/edit the annotation on an item by clicking the “note” icon in the item component
  - When a notepad item is selected, I can delete it from the keyboard with Cmd/Meta + Backspace
- I can delete all items in the notepad by clicking the “trash can” icon and confirming that I want to delete all items when the confirmation message pops up
- I can create a Notebook from items in the notepad by clicking the “Create Notebook” button
- I can select one or more individual notepad items by Cmd/Meta + right clicking on them
- I can select a range of notepad items by clicking on one notepad item and then holding Shift while clicking on subsequent item that is above or below the currently selected item
- I can collapse the annotation box of any item by clicking the “note” icon in the item header
- I can select the next/previous item with the up and down arrow keys

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)
