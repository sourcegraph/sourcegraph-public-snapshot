<style>
  .demo{
    display: table;
    width: 35%;
    margin: 0.5em;
    padding: 1rem 1rem;
    color: var(--text-color);
    border-radius: 4px;
    border: 1px solid var(--sidebar-nav-active-bg);
    padding: 1rem;
    padding-top: 1rem;
    background-color: var(--sidebar-nav-active-bg);
  }
</style>

# Cody capabilities

Learn and understand what features Cody offers to ensure a streamlined code AI.

## Code chatbot

Cody is your AI-powered coding assistant that understands your entire codebase inside out. It goes beyond other AI code chatbots, aligning perfectly with your project's code conventions and architecture. You can chat with Cody right within your code editor or through the Sourcegraph sidebar.

Cody tells you which code files it reads to generate its response. In case of a wrong answer, please share feedback so we can improve it.

Examples of questions Cody can handle:

- How is our app's secret storage implemented on Linux?
- Where is the CI config for the web integration tests?
- Write a new GraphQL resolver for the AuditLog.
- Why is the UserConnectionResolver giving an error `unknown user`, and how do I fix it

<div class="getting-started">
  <a class="demo text-center" href="https://twitter.com/beyang/status/1647744307045228544">View Demo</a>
</div>

## Fix code inline

Cody can help you make interactive edits and refactor code by following natural-language instructions. To do so, add natural-language instructions to your code, select the relevant code, and run:

```bash
 Cody: Fixup(Ctrl+Opt+/) — for macOS

 Cody: Fixup(Ctrl+Alt+/) — for Windows
```

Cody will take it from there and figure out what edits to make.

Examples of fix-up instructions Cody can handle:

- Factor out any common helper functions (when multiple functions are selected)
- Use the imported CSS module's class `n`
- Extract the list item to a separate React component
- Handle errors better
- Add helpful debug log statements
- Make this work (and yes, it often does work—give it a try!)

<div class="getting-started">
  <a class="demo text-center" href="https://twitter.com/sqs/status/1647673013343780864">View Demo</a>
</div>

## Cody Recipes

Cody can generate unit tests, documentation, and more, leveraging its full awareness of your codebase. You can access various helpful recipes to streamline your development process with just a few clicks.

Select the recipes tab in your editor or right click on a code section, then choose one of the `Ask Cody > ...` recipes. You'll find options such as:

- Explain code
- Generate unit test
- Detect code smells
- Generate `docstring`
- Generate release notes
- Improve variable names
- Translate to different language
- Summarize recent code changes

## Code Autocomplete

While typing, Cody provides real-time code auto-completion based on the context around your open files and file history. This predictive feature ensures a smoother coding experience.
