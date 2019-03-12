import pRetry from 'p-retry'
import puppeteer from 'puppeteer'
import { OperationOptions } from 'retry'

/**
 * Retry function with more sensible defaults for e2e test assertions
 *
 * @param fn The async assertion function to retry
 * @param options Option overrides passed to pRetry
 */
export const retry = (fn: (attempt: number) => Promise<any>, options: OperationOptions = {}) =>
    pRetry(fn, { factor: 1, ...options })

/**
 * Looks up an environment variable and parses it as a boolean. Throws when not
 * set and no default is provided, or if parsing fails.
 */
export function readEnvBoolean({
    variable: variable,
    defaultValue,
}: {
    variable: string
    defaultValue?: boolean
}): boolean {
    const value = process.env[variable]

    if (!value) {
        if (defaultValue === undefined) {
            throw new Error(`Environment variable ${variable} must be set.`)
        }
        return defaultValue
    }

    try {
        return Boolean(JSON.parse(value))
    } catch (e) {
        throw new Error(`Incorrect environment variable ${variable}=${value}. Must be truthy or not set at all.`)
    }
}

/**
 * Looks up an environment variable. Throws when not set and no default is
 * provided.
 */
export function readEnvString({ variable, defaultValue }: { variable: string; defaultValue?: string }): string {
    const value = process.env[variable]

    if (!value) {
        if (defaultValue === undefined) {
            throw new Error(`Environment variable ${variable} must be set.`)
        }
        return defaultValue
    }
    return value
}

export async function ensureLoggedIn({
    page,
    baseURL,
    email = 'test@test.com',
    username = 'test',
    password = 'test',
}: {
    page: puppeteer.Page
    baseURL: string
    email?: string
    username?: string
    password?: string
}): Promise<void> {
    await page.goto(baseURL)
    const url = new URL(await page.url())
    if (url.pathname === '/site-admin/init') {
        await page.type('input[name=email]', email)
        await page.type('input[name=username]', username)
        await page.type('input[name=password]', password)
        await page.click('button[type=submit]')
        await page.waitForNavigation()
    } else if (url.pathname === '/sign-in') {
        await page.type('input', username)
        await page.type('input[name=password]', password)
        await page.click('button[type=submit]')
        await page.waitForNavigation()
    }
}
