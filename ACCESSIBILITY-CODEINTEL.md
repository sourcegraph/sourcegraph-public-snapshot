Issue link:

https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=accessibility%2Cwcag%2F2.1%2Cwcag%2F2.1%2Fauditing%2C&title=%5BAccessibility%5D%3A+

# 1

[Accessibility Audit] Code Intel: Fuzzy file search through a repository

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to a repository in Sourcegraph.
- Press Cmd + K (or equivalent based on platform).
- Optional: Type a string to narrow down the files shown.
- Select file via keyboard navigation (up and down + Enter) or by clicking.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 2

[Accessibility Audit] Code Intel: Navigate through directory tree in a repository

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to a repository in Sourcegraph
- Optional: Show left sidebar if hidden.
- Optional: Select “Files” in the sidebar if it is not selected.
- Optional: Click on directories to drill down further, potentially scrolling to navigate to the desired directory / file.
- Click on a file name to open it and read the code (likely has syntax highlighting).

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 3

[Accessibility Audit] Code Intel: Search symbols for a repository

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to a repository in Sourcegraph.
- Optional: Show left sidebar if hidden.
- Optional: Select the Symbols tab if the Files tab is active.
- Search for a symbol in the symbols search bar.
- Scroll through results and click on a result.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 4

[Accessibility Audit] Code Intel: Going to a definition from a reference

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Hover over an identifier.
- Click on the Go to Definition button in the Code Intel Popover. [Optional: If you hover over the identifier in the definition, the button will be inactivated and the label will be “You are at the definition”]

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 5

[Accessibility Audit] Code Intel: Finding references for a definition

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Hover over an identifier.
- Click on the Find references button in the Code Intel Popover.
- Optional: Potentially increase height of the References panel. (Note: I wasn’t able to capture the cursor change in the screenshot. There is a very subtle shadow change too near the boundary, but it’s hard to see without zooming in.)
- Select repository and file of interest from the References panel.
- Click on code snippet to navigate to reference.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 6

[Accessibility Audit] Code Intel: Finding implementations for a definition

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Hover over an identifier.
- Click on the Find implementations button in the Code Intel Popover.
- Optional: Potentially increase height of the Implementations panel.
- Select repository and file of interest from the Implementations panel.
- Click on code snippet to navigate to reference.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 7

[Accessibility Audit] Code Intel: Inspecting LSIF uploads for a repository

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to a repository in Sourcegraph.
- Click on the 🧠 Code Intelligence button.
- Scroll through the list of uploads.
- Optional: Adjust Upload State filter.
- Optional: Search uploads for specific uploads.
- Click on > next to an upload to inspect it.
- Potential options:
  - Click the 🗑 Delete upload button.
  - Expand Dependencies section
    - Optional: Show dependents.
    - Optional: Search dependencies.
    - Click on > next to an upload to inspect it.
  - Expand Retention overview section.
    - Optional: Click Show matching only.
    - Optional: Type filter in “Search Matches…”
    - Click on a retention policy.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 8

[Accessibility Audit] Code Intel: Inspecting auto-indexing jobs for a repository

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to a repository in Sourcegraph.
- Click on the 🧠 Code Intelligence button.
- Click on Auto-indexing.
- Potential options:
  - Run:
    - [Optional: Adjust Index state filter]
    - Click Enqueue
  - Inspect:
    - [Optional: Change Index state filter.]
    - [Optional: Search indexes for specific ones.]
    - Click on > for an index.
    - Potential options:
      - Click on 🗑 Delete index button.
      - Inspect steps:
        - Expand step.
        - Expand Log output and potentially scroll to read it.
        - [Optional: Collapse Log output.]
        - Collapse step.
      - Click on > next to Upload.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 9

[Accessibility Audit] Code Intel: Inspecting LSIF uploads for an instance (site-admin)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to site admin settings in Sourcegraph.
- Click on Uploads in the 🧠 Code Intelligence section in the sidebar.
- Rest of the steps are the same as for [Inspecting LSIF uploads for a repository](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit#heading=h.1t2ar7qztcrk)

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 10

[Accessibility Audit] Code Intel: Inspect auto-indexing jobs for an instance (site-admin)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to site admin settings in Sourcegraph.
- Click on Auto-indexing in the 🧠 Code Intelligence section in the sidebar.
- Rest of the steps are the same as for [Inspecting auto-indexing jobs for a repository](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit#heading=h.nokk92k9dc2)

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 11

[Accessibility Audit] Code Intel: Editing auto-indexing policies for a repository

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to a repository in Sourcegraph
- Click on the 🧠 Code Intelligence button.
- Click on Configuration policies.
- Potential options:
  - Click on an existing policy.
  - Click on Create New Policy.
    - Fill out various fields in form and click Create Policy.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 12

[Accessibility Audit] Code Intel: Editing auto-indexing configuration for a repository

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to a repository in Sourcegraph.
- Click on the 🧠 Code Intelligence button.
- Click on Auto-indexing configuration.
- Enter configuration in text input.
- Click the Save Changes button.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 13

[Accessibility Audit] Code Intel: Editing auto-indexing policies for an instance (site-admin)

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

- Navigate to site admin settings in Sourcegraph.
- Click on Configuration in the 🧠 Code Intelligence section in the sidebar.
- Rest of the steps are the same as for [Editing auto-indexing policies for a repository.](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit#heading=h.f4aysc1m2vd0)

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 14

[Accessibility Audit] Code Intel: Read file content

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1kA6aVOAgID_uPm-d6uEC1DG6WHjJ8fPT0UAXQcML_KQ/edit?usp=sharing). Use this for further context.

Note: This is primarily aimed to assess the _consumption_ of content. As in, it should be possible to read the content of a file without any issues, irrelevant of if a user has a screen reader, magnifier, etc.

- Navigate to a file in Sourcegraph
- Read file contents

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)
