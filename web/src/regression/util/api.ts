/**
 * Provides convenience functions for interacting with the Sourcegraph API from tests.
 */

import { gql, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { GraphQLClient } from './GraphQLClient'
import { map, tap, retryWhen, delayWhen, take } from 'rxjs/operators'
import { zip, timer, concat, throwError, defer } from 'rxjs'
import { CloneInProgressError, ECLONEINPROGESS } from '../../../../shared/src/backend/errors'
import { isErrorLike } from '../../../../shared/src/util/errors'

/**
 * Wait until all repositories in the list exist.
 */
export async function waitForRepos(gqlClient: GraphQLClient, ensureRepos: string[]): Promise<void> {
    await zip(
        // List of Observables that complete after each repository is successfully fetched.
        ...ensureRepos.map(repoName =>
            gqlClient
                .queryGraphQL(
                    gql`
                        query ResolveRev($repoName: String!) {
                            repository(name: $repoName) {
                                mirrorInfo {
                                    cloned
                                }
                            }
                        }
                    `,
                    { repoName }
                )
                .pipe(
                    map(dataOrThrowErrors),
                    // Wait until the repository is cloned even if it doesn't yet exist.
                    // waitForRepos might be called immediately after adding a new external service,
                    // and we have no guarantee that all the repositories from that external service
                    // will exist when the add-external-service endpoint returns.
                    tap(({ repository }) => {
                        if (!repository || !repository.mirrorInfo || !repository.mirrorInfo.cloned) {
                            throw new CloneInProgressError(repoName)
                        }
                    }),
                    retryWhen(errors =>
                        concat(
                            errors.pipe(
                                delayWhen(error => {
                                    if (isErrorLike(error) && error.code === ECLONEINPROGESS) {
                                        // Delay retry by 2s.
                                        return timer(2 * 1000)
                                    }
                                    // Throw all errors other than ECLONEINPROGRESS
                                    throw error
                                }),
                                take(60) // Up to 60 retries (an effective timeout of 2 minutes)
                            ),
                            defer(() => throwError(new Error(`Could not resolve repo ${repoName}: too many retries`)))
                        )
                    )
                )
        )
    ).toPromise()
}

export async function ensureNoTestExternalServices(
    gqlClient: GraphQLClient,
    options: {
        kind: GQL.ExternalServiceKind
        uniqueDisplayName: string
        deleteIfExist?: boolean
    }
): Promise<void> {
    if (!options.uniqueDisplayName.startsWith('[TEST]')) {
        throw new Error(
            `Test external service name ${JSON.stringify(options.uniqueDisplayName)} must start with "[TEST]".`
        )
    }

    const externalServices = await getExternalServices(gqlClient, options)
    if (externalServices.length === 0) {
        return
    }
    if (!options.deleteIfExist) {
        throw new Error('external services already exist, not deleting')
    }

    for (const externalService of externalServices) {
        await gqlClient
            .mutateGraphQL(
                gql`
                    mutation DeleteExternalService($externalService: ID!) {
                        deleteExternalService(externalService: $externalService) {
                            alwaysNil
                        }
                    }
                `,
                { externalService: externalService.id }
            )
            .toPromise()
    }
}

export function getExternalServices(
    gqlClient: GraphQLClient,
    options: {
        kind?: GQL.ExternalServiceKind
        uniqueDisplayName?: string
    } = {}
): Promise<GQL.IExternalService[]> {
    return gqlClient
        .queryGraphQL(
            gql`
                query ExternalServices($first: Int) {
                    externalServices(first: $first) {
                        nodes {
                            id
                            kind
                            displayName
                            config
                            createdAt
                            updatedAt
                            warning
                        }
                    }
                }
            `,
            { first: 100 }
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ externalServices }) =>
                externalServices.nodes.filter(
                    ({ displayName, kind }) =>
                        (options.uniqueDisplayName === undefined || options.uniqueDisplayName === displayName) &&
                        (options.kind === undefined || options.kind === kind)
                )
            )
        )
        .toPromise()
}

export async function ensureTestExternalService(
    gqlClient: GraphQLClient,
    options: {
        kind: GQL.ExternalServiceKind
        uniqueDisplayName: string
        config: Record<string, any>
    }
): Promise<void> {
    if (!options.uniqueDisplayName.startsWith('[TEST]')) {
        throw new Error(
            `Test external service name ${JSON.stringify(options.uniqueDisplayName)} must start with "[TEST]".`
        )
    }

    const externalServices = await getExternalServices(gqlClient, options)
    if (externalServices.length > 0) {
        return
    }

    // Add a new external service if one doesn't already exist.
    const input: GQL.IAddExternalServiceInput = {
        kind: options.kind,
        displayName: options.uniqueDisplayName,
        config: JSON.stringify(options.config),
    }
    await gqlClient
        .mutateGraphQL(
            gql`
                mutation addExternalService($input: AddExternalServiceInput!) {
                    addExternalService(input: $input) {
                        kind
                        displayName
                        config
                    }
                }
            `,
            { input }
        )
        .toPromise()
}

export async function getUser(gqlClient: GraphQLClient, username: string): Promise<GQL.IUser | null> {
    const user = await gqlClient
        .queryGraphQL(
            gql`
                query User($username: String!) {
                    user(username: $username) {
                        __typename
                        id
                        username
                        displayName
                        url
                        settingsURL
                        avatarURL
                        viewerCanAdminister
                        siteAdmin
                        createdAt
                        emails {
                            email
                            verified
                        }
                        organizations {
                            nodes {
                                id
                                displayName
                                name
                            }
                        }
                    }
                }
            `,
            { username }
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ user }) => user)
        )
        .toPromise()
    return user
}

export async function deleteUser(
    gqlClient: GraphQLClient,
    username: string,
    mustAlreadyExist: boolean = true
): Promise<void> {
    let user: GQL.IUser | null
    try {
        user = await getUser(gqlClient, username)
    } catch (err) {
        if (mustAlreadyExist) {
            throw err
        } else {
            return
        }
    }

    if (!user) {
        if (mustAlreadyExist) {
            throw new Error(`Fetched user ${username} was null`)
        } else {
            return
        }
    }

    await gqlClient
        .mutateGraphQL(
            gql`
                mutation DeleteUser($user: ID!, $hard: Boolean) {
                    deleteUser(user: $user, hard: $hard) {
                        alwaysNil
                    }
                }
            `,
            { hard: false, user: user.id }
        )
        .toPromise()
}

export async function setUserSiteAdmin(gqlClient: GraphQLClient, userID: GQL.ID, siteAdmin: boolean): Promise<void> {
    await gqlClient
        .mutateGraphQL(
            gql`
                mutation SetUserIsSiteAdmin($userID: ID!, $siteAdmin: Boolean!) {
                    setUserIsSiteAdmin(userID: $userID, siteAdmin: $siteAdmin) {
                        alwaysNil
                    }
                }
            `,
            { userID, siteAdmin }
        )
        .toPromise()
}

export function currentProductVersion(gqlClient: GraphQLClient): Promise<string> {
    return gqlClient
        .queryGraphQL(
            gql`
                query SiteFlags {
                    site {
                        productVersion
                    }
                }
            `,
            {}
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ site }) => site.productVersion)
        )
        .toPromise()
}
