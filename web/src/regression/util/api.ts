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
                                        // Delay retry by 1s.
                                        return timer(1000)
                                    }
                                    // Throw all errors other than ECLONEINPROGRESS
                                    throw error
                                }),
                                take(10)
                            ),
                            defer(() => throwError(new Error(`Could not resolve repo ${repoName}: too many retries`)))
                        )
                    )
                )
        )
    ).toPromise()
}

export async function ensureExternalService(
    gqlClient: GraphQLClient,
    options: {
        kind: GQL.ExternalServiceKind
        uniqueDisplayName: string
        config: Record<string, any>
    }
): Promise<void> {
    const externalServices = await gqlClient
        .queryGraphQL(
            gql`
                query ExternalServices($first: Int) {
                    externalServices(first: $first) {
                        nodes {
                            displayName
                        }
                    }
                }
            `,
            { first: 100 }
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ externalServices }) => externalServices)
        )
        .toPromise()
    if (externalServices.nodes.some(({ displayName }) => displayName === options.uniqueDisplayName)) {
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
