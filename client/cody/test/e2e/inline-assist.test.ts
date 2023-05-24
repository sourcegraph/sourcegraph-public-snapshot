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
    // Wait for the code lens to show up to ensure that the fixup has been applied
    // await expect(page.getByText('Processing')).toBeVisible()

    // Ensures Code Lens is added
    await expect(page.getByText('Edited by Cody')).toBeVisible()
    await expect(page.getByText('<title>Goodbye Cody</title>')).toBeVisible()

    // Ensures Cody's fixup is displayed in comment thread
    await expect(page.getByText('Check your document for updates from Cody.')).toBeVisible()

    // Ensures Decorations is displayed by checking hover text
    await page.getByText('>Goodbye Cody<').hover()
    // The decoration text on hover should start with 'Cody Fixup #' and end with random number
    await expect(page.locator('.hover-contents').first()).toHaveText(/Cody Fixup #\d+/)
})
