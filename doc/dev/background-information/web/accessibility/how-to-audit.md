# How to conduct an accessibility audit

## Preparation

Before starting your audit, you should ensure the following statements are true:

- A relevant GitHub issue has been created for this audit, and it is attached to the [WCAG 2.1 auditing tracking issue](https://github.com/sourcegraph/sourcegraph/issues/31475).
- The user journey you are about to audit feels appropriately scoped.
  - Example: Instead of auditing the entirety of code insights, the audit should be focused on a specific part, such as creating a new code insight.
- You have read and understood [How to use a screen reader](how-to-screen-reader.md).


## Auditing a user journey

1. Navigate through the journey using **only** the <kbd>tab</kbd> key.
    - Ensure that you are able to access all important **actions**.
    - Ensure that the current focus position is always **clear** and **visible**.
2. Enable a screen reader. Navigate through the user journey **without** looking at your screen.
    - Would a user be able to understand the **content** of the journey?
    - Would a user be able to correctly and predictably perform each important **action**?
    - Are headings used correctly? Would a user be able to predictably navigate between each heading within the journey?
    - Are images correctly described? Any graphical content that is important for the journey should be read out by the screen reader.
3. Run through the [detailed checklist](#detailed-checklist) and ensure that there are no major issues.

## Raising an issue

TODO
