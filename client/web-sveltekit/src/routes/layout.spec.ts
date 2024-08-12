import { test, expect, defineConfig } from '$testing/integration'

export default defineConfig({
    name: 'Root layout',
})

test('has sign in button', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('link', { name: 'Sign in' }).click()
    await expect(page).toHaveURL('/sign-in')
})

test('has user menu', async ({ sg, page }) => {
    await sg.signIn({ username: 'test' })
    const userMenu = page.getByLabel('Open user menu')

    await page.goto('/')
    await expect(userMenu).toBeVisible()

    // Open user menu
    await userMenu.click()
    await expect(page.getByRole('heading', { name: 'Signed in as @test' })).toBeVisible()
})

test.describe('cody top level navigation', () => {
    const topNavName = 'Cody'

    const tests = [
        {
            name: 'sourcegraph.com, signed out',
            context: { sourcegraphDotComMode: true },
            signedIn: false,
            expectedTopNav: '/cody',
        },
        {
            name: 'sourcegraph.com, signed in',
            context: { sourcegraphDotComMode: true },
            signedIn: true,
            expectedTopNav: '/cody/chat',
        },
        {
            name: 'enterprise, no user cody',
            context: { sourcegraphDotComMode: false, codyEnabledOnInstance: true, codyEnabledForCurrentUser: false },
            signedIn: true,
            expectedTopNav: '/cody/dashboard',
        },
        {
            name: 'enterprise, user cody',
            context: { sourcegraphDotComMode: false, codyEnabledOnInstance: true, codyEnabledForCurrentUser: true },
            signedIn: true,
            expectedTopNav: '/cody/chat',
        },
        {
            name: 'enterprise, no cody',
            context: { sourcegraphDotComMode: false, codyEnabledOnInstance: false, codyEnabledForCurrentUser: false },
            signedIn: true,
        },
    ]

    tests.forEach(({ name, context, signedIn, expectedTopNav }) => {
        test(name, async ({ sg, page }) => {
            const mainNav = page.getByLabel('Main')
            const topNavCodyEntry = mainNav.getByRole('link', { name: topNavName })
            await sg.setWindowContext(context)

            if (signedIn) {
                await sg.signIn({ username: 'test' })
            }

            await page.goto('/')
            await expect(mainNav).toBeVisible()

            if (typeof expectedTopNav === 'string') {
                await expect(topNavCodyEntry).toHaveAttribute('href', expectedTopNav)
            } else {
                await expect(topNavCodyEntry).not.toBeAttached()
            }
        })
    })
})

test('has global notifications', async ({ sg, page }) => {
    sg.mockTypes({
        Query: () => ({
            site: {
                id: 'test_001',
                freeUsersExceeded: true,
                needsRepositoryConfiguration: true,
                alerts: [
                    {
                        type: 'WARNING' as any,
                        isDismissibleWithKey: null,
                        message:
                            '[**Update external service configuration**](/site-admin/external-services) to resolve problems:\n* perforce provider "perforce.sgdev.org:1666": rpc error: code = InvalidArgument desc = exit status 1 (output follows)\n\nFailed client connect, server using SSL.\nClient must add SSL protocol prefix to P4PORT.\n\n* perforce provider "perforce.sgdev.org:1666": rpc error: code = InvalidArgument desc = exit status 1 (output follows)\n\nFailed client connect, server using SSL.\nClient must add SSL protocol prefix to P4PORT.\n',
                    },
                ],
                productSubscription: {
                    license: {
                        expiresAt: '2024-03-16T22:59:59Z',
                    },
                    noLicenseWarningUserCount: null,
                },
            },
        }),
    })

    await page.goto('/')

    const alerts = page.getByRole('alert')

    await expect(alerts.first()).toBeVisible()
    await expect(alerts).toHaveCount(4)
})

// Because of how SvelteKit routing works, having a URL with a file path that includes route segements is
// problematic. We solve this problem by automatically encoding file paths in the URL. This test ensures
// that this behavior works as expected.
test('automatic file path encoding', async ({ sg, page }) => {
    sg.mockOperations({
        ResolveRepoRevision(variables) {
            return {
                repositoryRedirect: {
                    id: '1',
                },
            }
        },
    })
    await page.goto('/sourcegraph/sourcegraph/-/blob/app/src/routes/-/blob/page.ts')
    // If this didn't work we would render a 'Error: Not found' page
    await expect(page.getByRole('heading', { name: 'sourcegraph/sourcegraph' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Error' })).not.toBeVisible()
})
