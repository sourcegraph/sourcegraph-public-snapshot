import { test, expect, defineConfig } from '../testing/integration'

export default defineConfig({
    name: 'Root layout',
})

test('has sign in button', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('link', { name: 'Sign in' }).click()
    await expect(page).toHaveURL('/sign-in')
})

test('has experimental web app toggle button', async ({ page }) => {
    await page.goto('/')
    await page.getByLabel('Leave experimental web app').click()
    await expect(page).toHaveURL(/\?feat=-web-next/)
})

test('has user menu', async ({ sg, page }) => {
    sg.signIn({ username: 'test' })
    const userMenu = page.getByLabel('Open user menu')

    await page.goto('/')
    await expect(userMenu).toBeVisible()

    // Open user menu
    await userMenu.click()
    await expect(page.getByRole('heading', { name: 'Signed in as @test' })).toBeVisible()
})
