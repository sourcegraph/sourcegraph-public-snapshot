import { expect } from '@playwright/test'

import { sidebarExplorer, sidebarSignin } from './common'
import { test } from './helpers'

test('task tree view for non-stop cody', async ({ page, sidebar }) => {
    // Sign into Cody
    await sidebarSignin(page, sidebar)

    // Open the Explorer view from the sidebar
    await sidebarExplorer(page).click()

    // Open the index.html file from the tree view
    await page.getByRole('treeitem', { name: 'index.html' }).locator('a').dblclick()

    // Bring the cody sidebar to the foreground
    await page.click('[aria-label="Sourcegraph Cody"]')

    // Expand the task tree view
    await page.getByRole('button', { name: 'Fixups Section' }).click()

    // Open the command palette by clicking on the Cody Icon
    // Expect to see fail to start because no text was selected
    await page.getByRole('button', { name: /Cody: Fixup.*/ }).click()
    await expect(page.getByText(/^Cody Fixups: Failed to start.*/)).toBeVisible()

    // Find the text hello cody, and then highlight the text
    await page.getByText('<title>Hello Cody</title>').click()

    // Hightlight the whole line
    await page.keyboard.down('Shift')
    await page.keyboard.press('ArrowDown')

    // Open the command palette by clicking on the Cody Icon
    await page.getByRole('button', { name: /Cody: Fixup.*/ }).click()
    // Type in the instruction for fixup
    await page.keyboard.type('replace hello with goodbye')
    // Press enter to submit the fixup
    await page.keyboard.press('Enter')

    // Expect to see the fixup instruction in the task tree view
    await expect(page.getByText('1 fixup, 1 ready')).toBeVisible()
    await expect(page.getByText('No pending Cody fixups')).not.toBeVisible()

    // Diff view button
    await page.locator('a').filter({ hasText: 'replace hello with goodbye' }).click()
    await page.getByRole('button', { name: 'Cody: Show diff for fixup' }).click()
    await expect(page.getByText(/^Diff view for task.*/)).toBeVisible()

    // Apply fixup button on Click
    await page.locator('a').filter({ hasText: 'replace hello with goodbye' }).click()
    await page.getByRole('button', { name: 'Cody: Apply fixup' }).click()
    await expect(page.getByText(/^Applying fixup for task.*/)).toBeVisible()

    // Close the file tab and then clicking on the tree item again should open the file again
    await page.getByRole('button', { name: /^Close.*/ }).click()
    await expect(page.getByText('<title>Hello Cody</title>')).not.toBeVisible()
    await page.locator('a').filter({ hasText: 'replace hello with goodbye' }).click()
    await expect(page.getByText('<title>Hello Cody</title>')).toBeVisible()

    // Collapse the task tree view
    await page.getByRole('button', { name: 'Fixups Section' }).click()
    await expect(page.getByText('replace hello with good bye')).not.toBeVisible()

    // The chat view should be visible again
    await expect(sidebar.getByText(/^Check your doc.*/)).toBeVisible()
})
