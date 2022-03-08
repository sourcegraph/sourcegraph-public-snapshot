# How to conduct an accessibility audit

## Prerequisites

Before starting your audit, you should ensure the following statements are true:

- A relevant GitHub issue has been created for this audit, and it is attached to the [WCAG 2.1 auditing tracking issue](https://github.com/sourcegraph/sourcegraph/issues/31475).
- The user journey you are about to audit feels appropriately scoped.
  - Example: Instead of auditing the entirety of code insights, the audit should be focused on a specific part, such as creating a new code insight.
- You have read and understood [How to use a screen reader](how-to-screen-reader.md).

<i>If you have any issues, please contact the Frontend Platform team.</i>>

## Auditing a user journey

1. Navigate through the journey using **only** the <kbd>tab</kbd> key.
    - Ensure that you are able to access all important **actions**.
    - Ensure that the current focus position is always **clear** and **visible**.
2. Enable a screen reader. Navigate through the user journey **without** looking at your screen.
    - Would a user be able to understand the **content** of the journey?
    - Would a user be able to correctly and predictably perform each important **action**?
    - **Note:** Use the cheatsheet in [How to use a screen reader](how-to-screen-reader.md) to help you navigate.
3. Navigate through the user journey using a viewport that has a **width of 320px**.
    - Would a user be able to sufficiently read all required content in the journey?
    - Would a user be able to correctly and predictably perform each important **action**?
      - Keep in mind that it is typically harder to select small buttons and icons using a touch device.
    - **Note:** You don't need to use a physical mobile device to test this. Most browsers support simulating a mobile viewport - just ensure it is set to **320px**.
      - [Chrome documentation](https://developer.chrome.com/docs/devtools/device-mode/#viewport)
      - [Firefox documentation](https://developer.mozilla.org/en-US/docs/Tools/Responsive_Design_Mode)
4. Work through relevant sections from the [detailed checklist](detailed-checklist.md) and ensure that there are no issues.

## Raising an issue

TODO
