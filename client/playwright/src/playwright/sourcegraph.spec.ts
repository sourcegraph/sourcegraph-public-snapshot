import { test, expect } from '@playwright/test'

test('has title', async ({ page }) => {
    await page.goto('https://sourcegraph.com/search')

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Sourcegraph/)
})

test('get Batch Changes link', async ({ page }) => {
    await page.goto('/search')
    // await page.goto('/search')

    await page.getByText('About Sourcegraph').click();

    // await page.screenshot({ path: 'screenshot2.png' });

    // Expects page to have a heading with the name of Installation.
    await expect(page).toHaveURL('/')

})
