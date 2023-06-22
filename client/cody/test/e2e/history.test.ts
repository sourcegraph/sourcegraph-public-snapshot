import { expect } from '@playwright/test'

import { sidebarSignin } from './common'
import { test } from './helpers'

test.skip('checks for the chat history and new session', async ({ page, sidebar }) => {
    // Sign into Cody
    await sidebarSignin(page, sidebar)

    await page.click('[aria-label="Chat History"]')
    await expect(sidebar.getByText('Chat History')).toBeVisible()

    // start a new chat session and check history

    await page.click('[aria-label="Start a New Chat Session"]')
    await expect(sidebar.getByText("Hello! I'm Cody. I can write code and answer questions for you.")).toBeVisible()

    await sidebar.getByRole('textbox', { name: 'Text area' }).fill('Hello')
    await sidebar.locator('vscode-button').getByRole('img').click()
    await expect(sidebar.getByText('Hello')).toBeVisible()
    await page.getByRole('button', { name: 'Chat History' }).click()
    await sidebar.locator('vscode-button').filter({ hasText: 'Hello' }).click()
    await page.getByRole('button', { name: 'Chat History' }).click()
    await sidebar.locator('vscode-button').filter({ hasText: 'Hello' }).locator('i').click()
    await expect(sidebar.getByText('Hello')).not.toBeVisible()
})
