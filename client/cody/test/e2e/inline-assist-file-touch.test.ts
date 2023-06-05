import { expect } from '@playwright/test'

import { sidebarExplorer, sidebarSignin } from './common'
import { test } from './helpers'

test('start a fixup job from inline assist with valid auth', async ({ page, sidebar }) => {
    // Sign into Cody
    await sidebarSignin(page, sidebar)

    // Open the Explorer view from the sidebar
    await sidebarExplorer(page).click()

    // Open the index.html file from the tree view
    await page.getByRole('treeitem', { name: 'index.html' }).locator('a').dblclick()

    // wait for the editor to load
    await expect(page.getByText('<title>Hello Cody</title>')).toBeVisible()

    // Click on the doc and then the gutter to highlight whole line 7
    await page.locator('.view-lines > div:nth-child(11)').click()
    await page.getByText('7').click()

    // Click on line number 7 to open the comment thread
    await page.locator('.comment-range-glyph').nth(8).hover()
    await page.locator('.comment-range-glyph').nth(8).click()

    // After opening the comment thread, we need to wait for the editor to load
    await page.waitForSelector('.monaco-editor')
    await page.waitForSelector('.monaco-text-button')

    // Type in the instruction for fixup
    await page.keyboard.type('/touch replace hello with goodbye')
    // Click on the submit button with the name Ask Cody
    await page.click('.monaco-text-button')

    // Check if a new file called index.cody.html is created
    await expect(page.getByText('index.cody.html')).toBeVisible()

    // TODO check if content is correct. Currently blocked by ability to highlight in test
})
