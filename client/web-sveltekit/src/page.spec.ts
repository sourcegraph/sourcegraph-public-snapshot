import { test, expect } from './testing/integration'

test('smoke test: logo exists', async ({ page }) => {
    await page.goto('/search')
    const logo = page.getByRole('img', { name: 'Sourcegraph Logo' })
    await expect(logo).toBeVisible()
})
