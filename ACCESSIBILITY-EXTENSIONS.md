Issue link:

https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=accessibility%2Cwcag%2F2.1%2Cwcag%2F2.1%2Fauditing%2C&title=%5BAccessibility%5D%3A+

# 1

[Accessibility Audit] Integrations: Create extension

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Go to the extension registry page
- Click on the “+ Create extension button”
- Select publisher and input a name in the extension creation form

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 2

[Accessibility Audit] Integrations: Manage extension

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Go to a newly created extension that you own
- Click on the "Manage" tab
- Possible actions:
  - Update the name of the extension
  - Delete the extension
  - Publish the extension
    - Typically using the src-cli, we have an experimental in-browser version but we can remove this if we cannot make accessible.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 3

[Accessibility Audit] Integrations: Discover extensions

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Go to the extension registry page
- Enter a search query in the input element (placeholder text: “Search extensions…”).
- Execute the search.
- View extension results

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 4

[Accessibility Audit] Integrations: Filtering extensions

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Go to the extension registry page
- Use the category sidebar on the left to narrow the extensions that are displayed
- Use the "Show all" dropdown in the sidebar to select whether to display enabled, disabled, or all extensions, and whether to hide experimental extensions.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 5

[Accessibility Audit] Integrations: Deactivating/activating extensions

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Find an extension on the extension registry page
- Click the toggle switch to change the extension status between "Disabled" and "Enabled".

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 6

[Accessibility Audit] Integrations: Interacting with an extension on the search page

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Enable the [search-export](https://sourcegraph.com/extensions/sourcegraph/search-export) extension.
- Go to the homepage, execute a search.
- Observe the "Sourcegraph: Export search results" button in the search results info bar.
- Check that clicking the button triggers an action.
  - Note: We cannot guarantee WCAG 2.1 compliance for the actions of these extensions, as they can be created by third-parties.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 7

[Accessibility Audit] Integrations: Interacting with an extension on the file page (identifier)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Navigate to this file: https://sourcegraph.com/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/cmd/go-diff/go-diff.go
- Navigate to the "main()" identifier on line 23.
- View the tooltip that is rendered with additional information.
- Check that the "Find implementations" button from code intel extensions is rendered and triggers an action.
  - Note: We cannot guarantee WCAG 2.1 compliance for the actions of these extensions, as they can be created by third-parties.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 8

[Accessibility Audit] Integrations: Interacting with an extension on the file page (action item)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Navigate to this file: https://sourcegraph.com/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/cmd/go-diff/go-diff.go
- Observe the action items panel on the right side of the code view.
- Click the "Git blame" action item.
- Expect the "Git blame" action item to trigger an action.
  - Note: We cannot guarantee WCAG 2.1 compliance for the actions of these extensions, as they can be created by third-parties.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 9

[Accessibility Audit] Integrations: Interacting with an extension on the file page (line decoration)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Navigate to this file: https://sourcegraph.com/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/cmd/go-diff/go-diff.go
- Observe the action items panel on the right side of the code view.
- Click the "Git blame" action item.
- Navigate to a specific line in the file.
- Read the line decoration that is rendered with additional information.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 10

[Accessibility Audit] Integrations: Interacting with an extension on the file page (status bar)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Navigate to this file: https://sourcegraph.com/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/cmd/go-diff/go-diff.go
- Read the floating status bar on the bottom right of the code view.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 11

[Accessibility Audit] Integrations: Interacting with an extension on the file page (file decoration)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Enable the [codecov extension](https://sourcegraph.com/extensions/sourcegraph/codecov)
- Navigate to this repository: https://sourcegraph.com/github.com/sourcegraph/sourcegraph@684984c60f278fb223322c324e740e2ec0ce468a
- Observe the file decoration <meter> elements on both the persistent file tree and the tree page file tree.
  - Note: You might need to wait a few seconds for these to load.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 12

[Accessibility Audit] Integrations: Command palette

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1_IGekkIDuQugFA-2BYYWpYWpJy9O8PjiDrEEGmJxRRU/edit?usp=sharing). Use this for further context.

- Activate the command palette:
  - By clicking the icon in the navigation bar
  - By using the keyboard shortcut: Ctrl+P
- Use the command palette:
  - Filter commands by entering text into the input
  - Navigate the menu and select an item

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)
