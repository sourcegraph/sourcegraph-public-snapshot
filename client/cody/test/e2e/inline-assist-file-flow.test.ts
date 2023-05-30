import { expect } from '@playwright/test'

import { sidebarExplorer, sidebarSignin } from './common'
import { test } from './helpers'

test('start a fixup job from inline assist with valid auth', async ({ page, sidebar }) => {
    // Sign into Cody
    await sidebarSignin(page, sidebar)

    // Open the Explorer view from the sidebar
    await sidebarExplorer(page).click()

    // Select the index.html file from the tree view
    await page.getByRole('treeitem', { name: 'index.html' }).locator('a').dblclick()

    // Click on the gutter to open the comment thread
    await page.locator('div:nth-child(7) > .cldr').hover()
    await page.locator('div:nth-child(7) > .cldr').click()

    // After opening the comment thread, we need to wait for the editor to load
    await page.waitForSelector('.monaco-editor')
    await page.waitForSelector('.monaco-text-button')

    // Type in the instruction for fixup
    await page.keyboard.type('/touch replace hello with goodbye')
    // Click on the submit button with the name Ask Cody
    await page.click('.monaco-text-button')

    // Ensures a new file called index.cody.html is created with the new content
    await expect(page.getByText('index.cody.html')).toBeVisible()

    // Open the new file and check if the content is correct
    await page.getByRole('treeitem', { name: 'index.cody.html' }).locator('a').dblclick()
    await expect(page.getByText('<title>Goodbye Cody</title>')).toBeVisible()
})
