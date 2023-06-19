import { expect } from '@playwright/test'

import { SERVER_URL, VALID_TOKEN } from '../fixtures/mock-server'

import { test } from './helpers'

test('requires a valid auth token and allows logouts', async ({ page, sidebar }) => {
    await sidebar.getByRole('button', { name: 'Other Login Options...' }).click()
    await page.getByRole('option', { name: 'Sign in with URL and Access Token' }).click()
    await page.getByRole('combobox', { name: 'input' }).fill(SERVER_URL)
    await page.getByRole('combobox', { name: 'input' }).press('Enter')
    await page.getByRole('combobox', { name: 'input' }).fill('abcdefghijklmnopqrstuvwxyz')
    await page.getByRole('combobox', { name: 'input' }).press('Enter')

    await expect(sidebar.getByText('Invalid credentials')).toBeVisible()

    await sidebar.getByRole('button', { name: 'Other Login Options...' }).click()
    await page.getByRole('option', { name: 'Sign in with URL and Access Token' }).click()
    await page.getByRole('combobox', { name: 'input' }).fill(SERVER_URL)
    await page.getByRole('combobox', { name: 'input' }).press('Enter')
    await page.getByRole('combobox', { name: 'input' }).fill(VALID_TOKEN)
    await page.getByRole('combobox', { name: 'input' }).press('Enter')

    // Collapse the task tree view
    await page.getByRole('button', { name: 'Fixups Section' }).click()

    await expect(sidebar.getByText("Hello! I'm Cody.")).toBeVisible()

    // Check if embeddings server connection error is visible
    await expect(sidebar.getByText('Error while establishing embeddings server connection.')).not.toBeVisible()

    await page.getByRole('button', { name: 'Chat Section' }).hover()

    await page.getByRole('button', { name: 'More Actions...', exact: true }).click()
    await page.getByRole('button', { name: 'Sign out...' }).click()
    await page.getByRole('option', { name: 'sign-out http://localhost:49300, current' }).locator('a').click()

    await expect(sidebar.getByRole('button', { name: 'Other Login Options...' })).toBeVisible()
    await expect(sidebar.getByText('Invalid credentials')).not.toBeVisible()
})
