# Cody use cases

Cody can help with you write code and answers questions in several ways, including:

- [Writing code](#writing-code)
- [Testing code](#testing-code)
- [Fixing bugs](#fixing-bugs)
- [Onboarding to a new codebase](#onboarding-to-a-new-codebase)

## Writing code

When writing code, Cody can help by generating boilerplate code, refactoring code inline, or writing docstrings.

### Generate boilerplate code

Cody can generate boilerplate code to save you time. Use the Cody chat function to try this out.

For example:

- "Generate code for reading a text file in Go"
- "Write a new graphql resolver in Go for a new object type called PitaOrders"

### Refactor code with inline chat

Cody can also edit code directly in your file rather than making suggestions. You can prompt Cody to refactor code using inline chat.

To refactor code inline:

1. Highlight the snippet of code you'd like to refactor
2. Click the `+` icon to the left of the first line of code
3. In the Cody chat box, type `/edit` + your prompt for Cody. For example: `/edit refactor this code to be more easily readable`
4. Click `Ask Cody` (or alternatively use shortcut `Cmd + Enter`)

After submitting the request, Cody will prepare a code edit. You can click `Apply` and Cody will change the code inline, or you can click `Show Diff` to see Cody's proposed change.

Here are some examples of what you can do with inline /edit commands:

- "/edit Add a link here to the admin settings page"
- "/edit Make this code less verbose"
- "/edit Factor out any common helper functions"
- "/edit Convert this code to a react functional component"

Note: inline chat is currently only available in the VS Code extension.

## Fixing bugs

When debugging code, you can use Cody's natural language chat plus the fixup function to quickly isolate and fix issues.

### Ask Cody for debugging help

You can ask Cody for debugging help by pasting code snippets directly into the chat interface.

For example:

- Type "What's wrong with this code?" into the chat interface
- Create a space with `shift`+`enter`, then paste in a code snippet
- Submit the query

Cody will respond with potential issues it sees in your code. You can then respond by asking Cody to fix listed issues.

### Fix bugs with inline chat

The same inline chat `/edit` functionality that can be used to refactor code can also be used to debug, fix, and update code from directly within a file.

For example, try: "/edit fix the bug in this function"

## Onboarding to a new codebase

When onboarding to a new codebase, understanding code can be time consuming. Use Cody to help understand code faster and get oriented to where things are defined.

### Explain code at a high level

The `Explain Code` command provides concise explanations of what a file or code selection is doing. Use this when jumping into a new code file and trying to understand what's happening.

To use the command in the VS Code extension:

1. Highlight a code snippet
2. Open Cody code, and type `/explain`

To use the recipe in the IntelliJ extension:

1. Highlight a code snippet
2. Open the Cody "Recipes" panel
3. Click the "Explain selected code" recipe

### Show where things are defined

You can use the Cody chat sidebar to ask Cody to point out where things are defined. Cody will read files in your codebase to find them.

Try asking Cody: "Where is _X_ defined?"

For example: "Where is authentication defined?"

### Retrieve information from docs

You can also ask Cody to retrieve helpful information right from your docs.

- "What do our docs say about retrieving user data from Postgres?"
- "What do our docs say about authenication?"
- "What do our docs say about SAML?"

### Find resources that are already in use

You can ask Cody to find resources, such as libraries, that are already in use and available in a codebase.

Try asking Cody: "Do we already use any _language_ libraries that provide _XYZ_?"

For example: "Do we already use any Go libraries that provide an LRU cache? Where is that library used?"
