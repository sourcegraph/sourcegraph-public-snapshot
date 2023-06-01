# Non-Stop Cody

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>
</aside>

<iframe class="w-100" src="https://www.youtube.com/embed/cJj2ozoug60" title="Cody Inline Assist" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

<span class="badge badge-experimental">Experimental</span>  Non-Stop Cody tries to solve many development needs in real-time as you code, developed with one main goal: 

**preserve developer flow**

Like many other AI assistants, Cody is built to empower developers, which means:
- You never have to slow down for Cody
- You can understand what Cody is doing at any time through ambient signals
- When Cody performs poorly it is easy to recover

The seamless integration of the following Non-Stop Cody features allows developers to continue their work uninterrupted while using Cody, preserving their workflow and maximizing productivity.

## ✨ Cody Fixup

Sprinkles your code with instructions inline using natural language with `Cody Fixup`!

### How to use

1. Enter your instructions for Cody in your file
2. Highlight the code, including your instructions
3. Run `Cody: Fixup` (<kbd>Ctrl+Alt+/</kbd>/<kbd>Ctrl+Opt+/</kbd>). 
4. Leave the highlighted code untouched and wait for Cody to update your file, Cody will figure out what edits to make and will make it for you!

Examples:
- "Factor out any common helper functions" (when multiple functions are selected)
- "Use the imported CSS module's class names"
- "Extract the list item to a separate React component"
- "Handle errors better"
- "Add helpful debug log statements"
- "Make this work" (seriously, it often works--try it!)

## 🔮 Cody Inline Assist

Cody Inline Assist handles inline chat experience in VS Code using the CommentController from the VS Code API, which allows users to interact with Cody through comment threads in file.
 
Inline Assist provides the following functionalities:
- Create `Inline Chat` by creating a new comment thread with highlighted code.
- Post replies to a comment thread
- Run [/fix](#inline-fix) mode to perform inline edits
- Run [/touch](#inline-touch) mode to perform file creations with relavent contents

### Enable Inline Assist

While in experimental state, Cody Inline Assist needs to be enabled manually in your VS Code Settings.

1. Make sure your [Cody AI by Sourcegraph](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai) extension is on the latest version
   - <kbd>shift</kbd>+<kbd>cmd</kbd>+<kbd>x</kbd> to see all extensions, select Cody, confirm the version
2. Go to the `Cody Extension Settings`
   - Check the box for `Cody > Experimental Inline`
3. Restart or Reload VS Code

### How to use

1. Highlight the code where you need Cody's assists
2. Hover on the line number for the highlighted code in the left gutter 
3. Click on the `+` symbol to start a new Inline Chat with Cody
   1. If you click on the `+` symbol without highlighting code in your file, Cody will only looks at the line number you clicked and a few surrounding lines for context
4. Type in your questions for Cody in the Inline Chat input box
5. Click `Ask Cody` to submit your request
6. Cody will add a respond under your request in the Inline Chat once it's ready
7. In the meantime, you can continue working around your codebase

##  🪄 Cody Inline Fix

Cody Inline Fix is basically [Cody Fixup](#cody-fixup) in disguise. However, it lets you enter your instructions for Cody in the Inline Chat box instead of within your code. It also allows Cody to directly perform the edits inline based on context from highlighted code, with decorations indicates the process.

### How to use

To use `Cody Inline Fix`, you must first enable Cody Inline Assist.

1. Highlight the code where you want the content for the new file to be based upon
2. Start a new Inline Chat by clicking on the `+` symbol next to the highlighted line numbers in the left gutter
3. In the chatbox, enter `/fix` or `/f`, follows up with an instruction on what you would like Cody to `fix`
4. 

## 👉 Cody Inline Touch

Like the standard `touch` command in UNIX/Linux operating system, `Cody Touch` is a feature available in Cody Inline Assist that allows users to create new empty file within the current working directory for Cody to update with relevant contents inline.

### How to use

To use `Cody Inline Touch`, you must first enable Cody Inline Assist.

1. Highlight the code where you want the content for the new file to be based upon
2. Start a new Inline Chat by clicking on the `+` symbol next to the highlighted line numbers in the left gutter
3. In the chatbox, enter `/touch` or `/t`, follows up with an instruction on what you would like Cody to generate or create in the new file.
4. A new file will be created within the same directory where you initiate Inline Chat from
   - If the word "test" is included in your instruction, Cody will create a test file automatically in the same directory. e.g `helpers.ts` => `helpers.test.ts`
   - If a file with the same name exists, Cody will automatically insert new content to the end of the existing file
   - For queries without the word "test," Cody will create a new file in the same directory with "cody" between the file name and the extension, allowing users to easily identify the newly generated files. eg. `index.ts` => `index.cody.ts`

Examples:
- /touch add 5 new tests case for this (with a function highlighted)
- /touch create a component named UserCard with required props of name, age and photoUrl


Coming soon: Create multiple files with relavent contents.
