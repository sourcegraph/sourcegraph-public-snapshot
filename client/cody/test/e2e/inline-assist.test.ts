import { expect } from '@playwright/test'

import { SERVER_URL, VALID_TOKEN } from '../fixtures/mock-server'

import { test } from './helpers'

test('start a fixup job from inline assist with valid auth', async ({ page, sidebar }) => {
    // Sign into Cody
    await sidebar.getByRole('textbox', { name: 'Sourcegraph Instance URL' }).fill(SERVER_URL)
    await sidebar.getByRole('textbox', { name: 'Access Token (docs)' }).fill(VALID_TOKEN)
    await sidebar.getByRole('button', { name: 'Sign In' }).click()

    await expect(sidebar.getByText("Hello! I'm Cody.")).toBeVisible()

    // Open the Explorer view
    await page.click('[aria-label="Explorer (⇧⌘E)"]')
    // Select the second element from the tree view, which is the index.html file
    await page.locator('.monaco-highlighted-label').nth(1).click()

    // Click on line number 6 to open the comment thread
    await page.locator('.comment-diff-added').nth(5).hover()
    await page.locator('.comment-diff-added').nth(5).click()

    // After opening the comment thread, we need to wait for the editor to load
    await page.waitForSelector('.monaco-editor')
    await page.waitForSelector('.monaco-text-button')

    // Type in the instruction for fixup
    await page.keyboard.type('/fix replace hello with goodbye')
    // Click on the submit button with the name Ask Cody
    await page.click('.monaco-text-button')

    // TODO: Capture processing state. It is currently to quick to capture the processing elements
    // await expect(page.getByText('Processing')).toBeVisible()

    // Wait for the code lens to show up to ensure that the fixup has been applied
    await expect(page.getByText('Edited by Cody')).toBeVisible()
    await expect(page.getByText('<title>Goodbye Cody</title>')).toBeVisible()

    // Bring the cody sidebar to the foreground
    await page.click('[aria-label="Sourcegraph Cody"]')
    // Ensure that we remove the hover from the activity icon
    await page.getByRole('heading', { name: 'Sourcegraph Cody: Chat' }).hover()

    // Logout
    await page.click('[aria-label="Cody: Settings"]')
    await sidebar.getByRole('button', { name: 'Logout' }).click()

    await expect(sidebar.getByRole('button', { name: 'Sign In' })).toBeVisible()
    await expect(sidebar.getByText('Invalid credentials')).not.toBeVisible()
})
