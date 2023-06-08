import { expect } from '@playwright/test'

import { sidebarExplorer, sidebarSignin } from './common'
import { test } from './helpers'

const DECORATION_SELECTOR = 'div.view-overlays[role="presentation"] div[class*="TextEditorDecorationType"]'

test('decorations from un-applied Cody changes appear', async ({ page, sidebar }) => {
    // Sign into Cody
    await sidebarSignin(page, sidebar)

    // Open the Explorer view from the sidebar
    await sidebarExplorer(page).click()

    // Open the index.html file from the tree view
    await page.getByRole('treeitem', { name: 'index.html' }).locator('a').dblclick()

    // Count the existing decorations in the file; there should be none.
    // TODO: When communication from the background process to the test runner
    // is possible, extract the FixupDecorator's decoration fields' keys and
    // select these exactly.
    const decorations = page.locator(DECORATION_SELECTOR)
    expect(await decorations.count()).toBe(0)

    // Find the text hello cody, and then highlight the text
    await page.getByText('<title>Hello Cody</title>').click()

    // Highlight the whole line
    await page.keyboard.down('Shift')
    await page.keyboard.press('ArrowDown')

    // Open the command palette by clicking on the Cody Icon
    await page.getByRole('button', { name: /Cody: Fixup.*/ }).click()

    // Wait for the input box to appear
    await page.getByPlaceholder('Ask Cody to edit your code, or use /chat to ask a question.').click()
    // Type in the instruction for fixup
    await page.keyboard.type('replace hello with goodbye')
    // Press enter to submit the fixup
    await page.keyboard.press('Enter')

    // Decorations should appear
    await page.waitForSelector(DECORATION_SELECTOR)

    // Extract the key of the decoration
    const decorationClassName = (await decorations.first().getAttribute('class'))
        ?.split(' ')
        .find(className => className.includes('TextEditorDecorationType'))
    expect(decorationClassName).toBeDefined()

    // Spray edits over where Cody planned to type to cause conflicts
    for (const ch of 'who needs titles?') {
        await page.keyboard.type(ch)
        await page.keyboard.press('ArrowRight')
    }

    // The decorations should change to conflict markers.
    await page.waitForSelector(`${DECORATION_SELECTOR}:not([class*="${decorationClassName}"])`)
})
