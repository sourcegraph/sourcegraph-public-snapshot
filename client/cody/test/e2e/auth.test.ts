import { expect } from '@playwright/test'

import { SERVER_URL, VALID_TOKEN } from '../fixtures/mock-server'

import { test } from './helpers'

test('requires a valid auth token and allows logouts', async ({ page, sidebar }) => {
    await sidebar.getByRole('textbox', { name: 'Sourcegraph Instance URL' }).fill(SERVER_URL)

    await sidebar.getByRole('textbox', { name: 'Access Token (docs)' }).fill('test token')
    await sidebar.getByRole('button', { name: 'Sign In' }).click()

    await expect(sidebar.getByText('Invalid credentials')).toBeVisible()

    await sidebar.getByRole('textbox', { name: 'Access Token (docs)' }).fill(VALID_TOKEN)
    await sidebar.getByRole('button', { name: 'Sign In' }).click()

    // Collapse the task tree view
    await page.getByRole('button', { name: 'Fixups Section' }).click()

    await expect(sidebar.getByText("Hello! I'm Cody.")).toBeVisible()

    // Check if embeddings server connection error is visible
    await expect(sidebar.getByText('Error while establishing embeddings server connection.')).not.toBeVisible()

    await page.getByRole('button', { name: 'Chat Section' }).hover()

    await page.click('[aria-label="Cody: Settings"]')
    await sidebar.getByRole('button', { name: 'Logout' }).click()

    await expect(sidebar.getByRole('button', { name: 'Sign In' })).toBeVisible()
    await expect(sidebar.getByText('Invalid credentials')).not.toBeVisible()
})
