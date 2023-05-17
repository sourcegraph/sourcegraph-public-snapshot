import { expect } from '@playwright/test'

import { test, getCodySidebar } from './helpers'

test('errors on invalid access token', async ({ page }) => {
    const sidebar = await getCodySidebar(page)

    await sidebar.getByRole('textbox', { name: 'Access Token (docs)' }).fill('test token')
    await sidebar.getByRole('button', { name: 'Sign In' }).click()

    await expect(sidebar.getByText('Invalid credentials')).toBeVisible()
})
