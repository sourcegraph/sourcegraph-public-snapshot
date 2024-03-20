import { test, expect } from '@playwright/test'

test('has title', async ({ page }) => {
    await page.goto(process.env.SOURCEGRAPH_BASE_URL!)

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Sourcegraph/)
})

test('get Batch Changes link', async ({ page }) => {
    await page.goto(process.env.SOURCEGRAPH_BASE_URL!)

    await page.getByText('Batch Changes').click()

    // Expects page to have a heading with the name of Installation.
    await expect(page.getByRole('heading', { name: 'Batch Changes' })).toBeVisible()
})
