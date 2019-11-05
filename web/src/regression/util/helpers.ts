import { GraphQLClient } from './GraphQLClient'
import { Driver } from '../../../../shared/src/e2e/driver'
import { gql, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import { catchError, map } from 'rxjs/operators'
import { throwError } from 'rxjs'
import { Key } from 'ts-key-enum'
import { deleteUser } from './api'
import { Config } from '../../../../shared/src/e2e/config'
import { ResourceDestructor } from './TestResourceManager'
import * as jsonc from '@sqs/jsonc-parser'
import * as jsoncEdit from '@sqs/jsonc-parser/lib/edit'
import * as puppeteer from 'puppeteer'
import {
    GitHubAuthProvider,
    GitLabAuthProvider,
    OpenIDConnectAuthProvider,
    SAMLAuthProvider,
} from '../../schema/critical.schema'

/**
 * Create the user with the specified password. Returns a destructor that destroys the test user.
 */
export async function ensureLoggedInOrCreateTestUser(
    driver: Driver,
    gqlClient: GraphQLClient,
    {
        username,
        deleteIfExists,
        testUserPassword,
    }: {
        username: string
        deleteIfExists?: boolean
    } & Pick<Config, 'testUserPassword'>
): Promise<ResourceDestructor> {
    const userDestructor = (): Promise<void> => deleteUser(gqlClient, username, false)

    if (!username.startsWith('test-')) {
        throw new Error(`Test username must start with "test-" (was ${JSON.stringify(username)})`)
    }

    if (deleteIfExists) {
        await deleteUser(gqlClient, username, false)
    } else {
        // Attempt to log in first
        try {
            await driver.ensureLoggedIn({ username, password: testUserPassword })
            return userDestructor
        } catch (err) {
            console.log(`Login failed (error: ${err.message}), will attempt to create user ${JSON.stringify(username)}`)
        }
    }

    await createTestUser(driver, gqlClient, { username, testUserPassword })
    await driver.ensureLoggedIn({ username, password: testUserPassword })
    return userDestructor
}

async function createTestUser(
    driver: Driver,
    gqlClient: GraphQLClient,
    { username, testUserPassword }: { username: string } & Pick<Config, 'testUserPassword'>
): Promise<void> {
    // If there's an error, try to create the user
    const passwordResetURL = await gqlClient
        .mutateGraphQL(
            gql`
                mutation CreateUser($username: String!, $email: String) {
                    createUser(username: $username, email: $email) {
                        resetPasswordURL
                    }
                }
            `,
            { username }
        )
        .pipe(
            map(dataOrThrowErrors),
            catchError(err =>
                throwError(new Error(`Could not create user ${JSON.stringify(username)}: ${err.message})`))
            ),
            map(({ createUser }) => createUser.resetPasswordURL)
        )
        .toPromise()
    if (!passwordResetURL) {
        throw new Error('passwordResetURL was empty')
    }

    await driver.page.goto(passwordResetURL)
    await driver.page.keyboard.type(testUserPassword)
    await driver.page.keyboard.down(Key.Enter)

    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    await driver.page.waitForFunction(() => document.body.textContent!.includes('Your password was reset'))
}

export async function clickAndWaitForNavigation(handle: puppeteer.ElementHandle, page: puppeteer.Page): Promise<void> {
    await Promise.all([handle.click(), page.waitForNavigation()])
}

/**
 * Navigate to the management console GUI and add the specified authentication provider. Returns a
 * function that restores the critical site config to its previous value (before adding the
 * authentication provider).
 */
export async function createAuthProviderGUI(
    driver: Driver,
    managementConsoleUrl: string,
    managementConsolePassword: string,
    authProvider: GitHubAuthProvider | GitLabAuthProvider | OpenIDConnectAuthProvider | SAMLAuthProvider
): Promise<ResourceDestructor> {
    const authHeaders = {
        Authorization: `Basic ${new Buffer(`:${managementConsolePassword}`).toString('base64')}`,
    }

    await driver.page.setExtraHTTPHeaders(authHeaders)
    await driver.goToURLWithInvalidTLS(managementConsoleUrl)

    const oldCriticalConfig = await driver.page.evaluate(async managementConsoleUrl => {
        const res = await fetch(managementConsoleUrl + '/api/get', { method: 'GET' })
        return (await res.json()).Contents
    }, managementConsoleUrl)
    const parsedOldConfig = jsonc.parse(oldCriticalConfig)
    const authProviders = parsedOldConfig['auth.providers'] as any[]
    if (
        authProviders.filter(p => p.type === authProvider.type && p.displayName === authProvider.displayName).length > 0
    ) {
        return () => Promise.resolve()
    }

    const newCriticalConfig = jsonc.applyEdits(
        oldCriticalConfig,
        jsoncEdit.setProperty(oldCriticalConfig, ['auth.providers', -1], authProvider, {
            eol: '\n',
            insertSpaces: true,
            tabSize: 2,
        })
    )
    await driver.replaceText({
        selector: '.monaco-editor',
        newText: newCriticalConfig,
        selectMethod: 'keyboard',
        enterTextMethod: 'paste',
    })
    await (await driver.findElementWithText('Save changes')).click()
    await driver.findElementWithText('Saved!', { wait: { timeout: 1000 } })
    await driver.page.setExtraHTTPHeaders({})

    return async () => {
        await driver.page.setExtraHTTPHeaders(authHeaders)
        await driver.goToURLWithInvalidTLS(managementConsoleUrl)

        await driver.replaceText({
            selector: '.monaco-editor',
            newText: oldCriticalConfig,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        await (await driver.findElementWithText('Save changes')).click()
        await driver.findElementWithText('Saved!', { wait: { timeout: 500 } })

        await driver.page.setExtraHTTPHeaders({})
    }
}
