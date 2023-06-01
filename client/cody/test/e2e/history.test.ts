import { expect } from '@playwright/test'

import { SERVER_URL, VALID_TOKEN } from '../fixtures/mock-server'

import { test } from './helpers'

test('checks for the chat history and new session', async ({ page, sidebar }) => {
    await sidebar.getByRole('textbox', { name: 'Sourcegraph Instance URL' }).fill(SERVER_URL)

    await sidebar.getByRole('textbox', { name: 'Access Token (docs)' }).fill(VALID_TOKEN)
    await sidebar.getByRole('button', { name: 'Sign In' }).click()

    // Collapse the task tree view
    await page.getByRole('button', { name: 'Fixups Section' }).click()

    await expect(sidebar.getByText("Hello! I'm Cody.")).toBeVisible()

    await page.getByRole('button', { name: 'Chat Section' }).hover()

    await page.click('[aria-label="Cody: Chat History"]')
    await expect(sidebar.getByText('Chat History')).toBeVisible()

    await page.click('[aria-label="Cody: Start a New Chat Session"]')
    await expect(sidebar.getByText("Hello! I'm Cody. I can write code and answer questions for you.")).toBeVisible()
})
