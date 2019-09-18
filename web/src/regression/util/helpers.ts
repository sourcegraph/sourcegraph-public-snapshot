import { GraphQLClient } from './GraphQLClient'
import { Driver } from '../../../../shared/src/e2e/driver'
import { gql, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import { catchError, map } from 'rxjs/operators'
import { throwError } from 'rxjs'
import { Key } from 'ts-key-enum'

/**
 * Create the user with the specified password
 */
export async function ensureLoggedInOrCreateUser({
    driver,
    gqlClient,
    username,
    password,
}: {
    driver: Driver
    gqlClient: GraphQLClient
    username: string
    password: string
}): Promise<void> {
    // Attempt to log in
    try {
        await driver.ensureLoggedIn({ username, password })
        return
    } catch (err) {
        console.log(`Login failed (error: ${err.message}), will attempt to create user ${JSON.stringify(username)}`)
    }

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
                throwError(
                    new Error(
                        `User likely alredy exists, but with a different password. Please delete user ${JSON.stringify(
                            username
                        )} and retry. (Underlying error: ${err.message})`
                    )
                )
            ),
            map(({ createUser }) => createUser.resetPasswordURL)
        )
        .toPromise()
    if (!passwordResetURL) {
        throw new Error('passwordResetURL was empty')
    }

    await driver.page.goto(passwordResetURL)
    await driver.page.keyboard.type(password)
    await driver.page.keyboard.down(Key.Enter)

    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    await driver.page.waitForFunction(() => document.body.textContent!.includes('Your password was reset'))
    await driver.ensureLoggedIn({ username, password })
}
