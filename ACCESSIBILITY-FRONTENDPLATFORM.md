Issue link:

https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=accessibility%2Cwcag%2F2.1%2Cwcag%2F2.1%2Fauditing%2C&title=%5BAccessibility%5D%3A+

# 1

[Accessibility Audit] Frontend Platform: Global navigation bar

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

General navigation:

- Go to Sourcegraph.com
- Navigate across the navigation bar
- Action an item and expect to be navigated

Navigation menu:

- Navigate to "Code search" in the navigation bar
- Open the dropdown and view contnets
- Action an item and expect to be navigated

User profile menu:

- As an authenticated user, navigate to the user profile menu.
- Open the menu and view contents
- Action an item and expect to be navigated

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 2

[Accessibility Audit] Frontend Platform: Global footer

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

General navigation:

- Go to Sourcegraph.com
- Navigate across the footer
- Action an item and expect to be navigated

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 3

[Accessibility Audit] Frontend Platform: Feedback forms

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Using feedback prompt:

- In the global navigation bar, select “Feedback”
- Enter some free-form text as feedback content
- Select a feedback rating
- Click “Submit”

Using feedback widget (NPS survey):

- Open Sourcegraph.com in an incognito window (or clear storage)
- In the browser console, run: localStorage.setItem('temporarySettings', '{"user.daysActiveCount":2}')
- This will ensure the feedback widget will trigger on the next page load
- Reload the page, view the feedback widget
- Enter a score
- Complete the survey form
- Click submit

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 4

[Accessibility Audit] Frontend Platform: Global keyboard shortcuts

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Note: Global keyboard shortcuts need some work to ensure they are accessible. See [this issue](https://github.com/sourcegraph/sourcegraph/issues/34410).

Viewing global shortcuts:

- As an authenticated user, open the user profile menu.
- Select “Keyboard shortcuts”
- OR: Cmd + ? to open the same view

Currently unimplemented journeys. See [this issue](https://github.com/sourcegraph/sourcegraph/issues/34410):

- Disabling keyboard shortcuts
- Remapping keyboard shortcuts

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 5

[Accessibility Audit] Frontend Platform: Color theme

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Switching color theme:

- As an authenticated user, open the user profile menu.
- Use the “Theme” dropdown to switch between color themes

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 6

[Accessibility Audit] Frontend Platform: Repositories selector

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Selecting a different repository:

- Go to the repositories header
  - (Optional) Click the repository name button (e.g. sourcegraph/sourcegraph) to go to the root page of the repository
- Click the dropdown arrow to trigger the popover
- View listed repositories
- Select “Show more” to load more results
- Focus the input and enter text to search for specific results
- Select a repository to navigate to it

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 7

[Accessibility Audit] Frontend Platform: Branch/tag/commit selector

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Selecting a different revision:

- Go to the branch/tag/commit selector
- View the current revision
- Click the selector to trigger the popover
- View listed revisions
- Select “Show more” to load more results
- Focus the input and enter text to search for specific results
- Select a revision to navigate to it
- Change tab to filter between revision types

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 8

[Accessibility Audit] Frontend Platform: File path

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Navigating the file path:

- Go to a specific file
- Click on a specific file/folder to navigate to it
- Click the copy icon to copy the path to your clipboard
- Click on the first “/” to navigate to the repository root

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 9

[Accessibility Audit] Frontend Platform: Tree page (pre-redesign)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Notes:

- Enable feature flag: “new-repo-page”
- This page does not include the file/symbol sidebar.

General navigation - Tabs

- Go to tree page
- Navigate between top level tabs (e.g. Commits, Branches)
- Click on a tab to navigate to the relevant page

View file tree

- Go to tree page
- Navigate to “Files and directories”
- Navigate through the file tree
- Click on a file path to navigate to said path

View recent commits

- Go to tree page
- Navigate to “Changes”
- View commit details
  - Click on the “...” icon to expand long commit messages
- Click on the commit hash to navigate to that commit diff

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 10

[Accessibility Audit] Frontend Platform: Tree page (post-redesign)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Notes:

- Disable feature flag: “new-repo-page”
- This page does not include the file/symbol sidebar.

General navigation - Tabs

- Go to tree page
- Navigate between top level tabs (e.g. Commits, Branches)
- Click on a tab to navigate to the relevant page

View README

- Go to tree page
- README content should be displayed if available
  - Note: This is external content so we are limited in what we can control here. We should avoid modifying this if possible.
    - Example: We might have to accept multiple <H1> elements on this page, as one might be provided in the README. We cannot reliably shift headings down.

View recent commits

- Go to tree page
- Navigate to “Recent commits”
- Click on the commit hash to navigate to that commit diff
- Click “Show more” to load more commits

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 11

[Accessibility Audit] Frontend Platform: Commits page

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Note: This page doesn’t have significant changes after the repository redesign. Double check this is still the case before actioning these user journeys in any way.

View commits:

- Go to commits tab
- Click on a commit message to navigate to that commit diff
  - Click on the “...” icon to expand long commit messages
- Click on a commit hash to navigate to that commit diff
- Click on the copy icon to copy the fully commit SHA hash to your clipboard
- Click on the files icon to navigate back to the repository tree page with the selected commit used as the repository revision.
- Click “Show more” to load more commits

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 12

[Accessibility Audit] Frontend Platform: Branches page

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Note: This page doesn’t have significant changes after the repository redesign. Double check this is still the case before actioning these user journeys in any way.

View branches - Overview:

- Go to branches tab
- View the default branch
- View active branches
- Click on a branch to navigate back to the repository tree page with the selected branch used as the repository revision.
- Click “View more branches” to navigate to the “All branches” tab (See below)

View branches - All branches:

- Go to branches tab
- Click “All branches”
- Click on a branch to navigate back to the repository tree page with the selected branch used as the repository revision.
- Focus the input and enter text to search for specific branches
- Select “Show more” to load more branches

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 13

[Accessibility Audit] Frontend Platform: Tags page

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Note: This page doesn’t have significant changes after the repository redesign. Double check this is still the case before actioning these user journeys in any way.

View tags:

- Go to tags tab
- Click on a tag to navigate back to the repository tree page with the selected tag used as the repository revision.
- Focus the input and enter text to search for specific tags
- Select “Show more” to load more tags

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 14

[Accessibility Audit] Frontend Platform: Compare page

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Notes:

- This page doesn’t have significant changes after the repository redesign. Double check this is still the case before actioning these user journeys in any way.
- Do not audit the diff view on this page, we have a separate issue for that

Selecting comparison revisions:

- Follow Branch/Tag/Commit selector journey
- Additional: Provide custom Git Revspec for speculative search
  s- Example: Change base to main^

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 15

[Accessibility Audit] Frontend Platform: Contributors page

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Notes:

- This page doesn’t have significant changes after the repository redesign. Double check this is still the case before actioning these user journeys in any way.
- This page has some bugs. [Issue to address](https://github.com/sourcegraph/sourcegraph/issues/34352)

Viewing contributors

- Go to contributors tab
- View list of contributors
- Click on the contributors latest commit message to navigate to that commit diff
- Click on the contributors commit count to view a diff search of all of their contributions

Filtering contributions

- Go to contributors tab
- Filter by time period
- Filter by revision range
- Filter by path

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 16

[Accessibility Audit] Frontend Platform: Repository settings (Admin)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Viewing settings

- Access settings page
  - Pre repository redesign: Access the settings tab on tree page
  - Post repository redesign: Access the settings action in the top right
- Navigate through the settings sidebar to view different pages
  - Note: We are not responsible for the content of these pages, just the navigation.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 17

[Accessibility Audit] Frontend Platform: User settings - Core

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Note: We own the core experience of the user settings page, we do not own each specific setting page - rather the navigation between those pages.

Viewing settings

- Access settings page
- Navigate between different setting tabs (e.g. Settings, Saved searches)
- Navigate through the settings sidebar to view different pages
  - Note: We are not responsible for the content of these pages, just the navigation.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 18

[Accessibility Audit] Frontend Platform: Site admin - Core

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Note: We own the core experience of the site admin page, we do not own each specific setting page - rather the navigation between those pages.

Viewing site admin

- Access site admin page
- Navigate through the site admin sidebar to view different pages
  - Note: We are not responsible for the content of these pages, just the navigation.
- Collapse sections of the sidebar to improve navigation

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 19

[Accessibility Audit] Frontend Platform: Commit diff view

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Note: We are not responsible for code view within the diff view. Specifically, we are not responsible for syntax highlighting (we are responsible for diff highlighting), hover definitions or any other code intel functionality.

Viewing diff commit metadata

- Navigate to commit diff view
- View commit author, message and date
- Click the copy icon next to the commit to copy the full commit SHA hash to your clipboard
- Click the copy icon next to the parent commit to copy the full commit SHA hash to your clipboard
- Click the parent commit to navigate to the commit diff view for said commit
- Click “Browse files at …” to navigate to the repository tree page with commit as the selected repository revision.

Viewing diff commit changes

- Click “Unified” or “Split” to toggle between different diff views
- Collapse changes to a file by clicking the down arrow next to a file path
- Click “View” next to the file path, to load that file in the typical code view, with the current commit as the selected revision.
- Click on the base commit above a file to navigate to the commit diff view for said commit
- Click on a line number to navigate directly to that line (useful for generating links to send)

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 20

[Accessibility Audit] Frontend Platform: Non-code file view

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1rSp77P0UDY6ewHq6iXmW0nuilBtkMW05n1ciMB8gQ3w/edit?usp=sharing). Use this for further context.

Viewing file content:

- Navigate to file
- Content should be displayed if available
- Click “Raw” or “Formatted” in the code header action bar to toggle between different view modes of the file content. (If available)
- Markdown files:
  - Note: This is external content so we are limited in what we can control here. We should avoid modifying this if possible.
    - Example: We might have to accept multiple <H1> elements on this page, as one might be provided in the README. We cannot reliably shift headings down.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)
