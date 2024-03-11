import { test, expect } from '../../testing/integration'

test('search input is autofocused', async ({ page }) => {
    await page.goto('/search')
    const searchInput = page.getByRole('textbox')
    const suggestions = page.getByLabel('Narrow your search')
    await expect(searchInput).toBeFocused()

    // Doesn't show suggestions without user interaction
    await expect(suggestions).not.toBeVisible()

    await searchInput.click()
    await expect(suggestions).toBeVisible()
})

test('shows suggestions', async ({ sg, page }) => {
    await page.goto('/search')
    const searchInput = page.getByRole('textbox')
    await searchInput.click()

    // Default suggestions
    await expect(page.getByLabel('Narrow your search')).toBeVisible()

    sg.mockTypes({
        SearchResults: () => ({
            repositories: [{ name: 'github.com/sourcegraph/sourcegraph' }],
            results: [
                {
                    __typename: 'FileMatch',
                    file: {
                        path: 'sourcegraph.md',
                        url: '',
                    },
                },
            ],
        }),
    })

    // Repo suggestions
    await searchInput.fill('source')
    await expect(page.getByLabel('Repositories')).toBeVisible()
    await expect(page.getByLabel('Files')).toBeVisible()

    // Fills suggestion
    await page.getByText('github.com/sourcegraph/sourcegraph').click()
    await expect(searchInput).toHaveText('repo:^github\\.com/sourcegraph/sourcegraph$ ')
})

test('submits search on enter', async ({ page }) => {
    await page.goto('/search')
    const searchInput = page.getByRole('textbox')
    await searchInput.fill('source')

    // Submit search
    await searchInput.press('Enter')
    await expect(page).toHaveURL(/\/search\?q=.+$/)
})

test('fills search query from URL', async ({ page }) => {
    await page.goto('/search?q=test')
    await expect(page.getByRole('textbox')).toHaveText('test')
})

test('main navbar menus are visible above search input', async ({ page, sg }) => {
    const dispatch = sg.mockSearchResults()
    await page.goto('/search?q=test')
    await dispatch()
    await page.getByRole('button', { name: 'Code Search' }).click()
    await page.getByRole('link', { name: 'Search Home' }).click()
    await expect(page).toHaveURL(/\/search$/)
})
