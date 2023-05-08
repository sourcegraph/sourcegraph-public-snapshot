# Cody use cases

Cody can help with you write code and answers questions in several ways, including:

- [Writing code](#writing-code)
- [Testing code](#testing-code)
- [Fixing bugs](#fixing-bugs)
- [Onboarding to a new codebase](#onboarding-to-a-new-codebase)

## Writing code

When writing code, Cody can help by generating boilerplate code, writing code with the fixup feature, or writing docstrings.

### Generate boilerplate code

Cody can generate boilerplate code to save you time. Use the Cody chat function to try this out.

For example:

- "Generate boilerplate code for reading a text file in Go"
- "Write a new graphql resolver in Go for a new object type called PitaOrders"

### Write code with fixup

Cody can also add and edit code directly in your file rather than making suggestions. This functionality is called "fixup," and it's designed to keep you in flow writing code alongside Cody.

To fixup some code, try:

1. Write something you'd like Cody to do inline in your file
2. Highlight this code plus any related code you'd like changed, then use the `ctrl`+`alt`+`/` hotkey (or the `Cody: Fixup` VS Code command)
3. Cody will work directly in your file

For example:

- "Add a link here to the admin settings page"
- "Add boilerplate code for a Java loop here"
- "Factor out any common helper functions"
- "Convert this code to a react functional component"

### Write docstrings

Cody can write docstrings for you to save time. To try this:

1. Highlight a code snippet, such as a function or class
2. Right click -> `Ask Cody` -> `Ask Cody: Generate Docstring`
3. Cody will provide a copy of the code with docstrings inserted

## Testing code

### Generate unit tests

When testing code, use Cody to quickly generate unit tests. Simply:

1. Highlight a component you want to test (such as a specific function)
2. Right click -> `Ask Cody` -> `Ask Cody: Generate Unit Test`
3. Cody provides code for a unit test

## Fixing bugs

When debugging code, you can use Cody's natural language chat plus the fixup function to quickly isolate and fix issues.

### Ask Cody for debugging help

You can ask Cody for debugging help by pasting code snippets directly into the chat interface.

For example:

- Type "What's wrong with this code?" into the chat interface
- Create a space with `shift`+`enter`, then paste in a code snippet
- Submit the query

Cody will respond with potential issues it sees in your code. You can then respond by asking Cody to fix listed issues.

### Fix bugs with fixup

The same fixup functionality that can be used to write code can also be used to debug, fix, and update code from directly within a file.

For example:

- Type "fix the bug in this function" directly underneath some function code
- Highlight the function plus the line you just typed
- Use the `ctrl`+`alt`+`/` hotkey (or the `Cody: Fixup` VS Code command)
- Cody will read your command and deploy its fix inline

## Onboarding to a new codebase

When onboarding to a new codebase, understanding code can be time consuming. Use Cody to help understand code faster and get oriented to where things are defined.

### Explain code at a high level

The `Explain selected code (high level)` recipe provides concise explanations of what a file or code selection is doing. Use this when jumping into a new code file and trying to understand what's happening.

You can use the recipe in 2 ways:

1. Highlight a code snippet and click the recipe. Cody will expain what is happening within that code.
2. Open a code file and click the recipe. Cody will explain the entire file.

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
