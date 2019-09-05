/**
 * Provides convenience functions for interacting with the Sourcegraph API from tests.
 */

import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { GraphQLClient } from './api'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'

/**
 * Block until all repositories in the list exist
 */
export async function waitForRepos(gqlClient: GraphQLClient, ensureRepos: string[]): Promise<void> {
    for (const repoName of ensureRepos) {
        await retry(async () => {
            const res = await gqlClient
                .queryGraphQL(
                    gql`
                        query ResolveRev($repoName: String!, $rev: String!) {
                            repository(name: $repoName) {
                                mirrorInfo {
                                    cloned
                                }
                            }
                        }
                    `,
                    { repoName, rev: '' }
                )
                .toPromise()
            if (
                res.data &&
                res.data.repository &&
                res.data.repository.mirrorInfo &&
                res.data.repository.mirrorInfo.cloned
            ) {
                return
            }
            throw new Error(`Repository ${repoName} did not exist.`)
        })
    }
}

export async function ensureExternalService(
    gqlClient: GraphQLClient,
    options: {
        kind: GQL.ExternalServiceKind
        uniqueDisplayName: string
        config: Record<string, string | number | boolean>
    }
): Promise<void> {
    const externalServicesRes = await gqlClient
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
        .toPromise()
    if (!externalServicesRes.data) {
        throw new Error('no `data` field found in GQL response')
    }
    const res = externalServicesRes.data.externalServices
    const existingMatches = res.nodes.filter(
        externalService => externalService.displayName === options.uniqueDisplayName
    )
    if (existingMatches.length > 0) {
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
                        id
                        kind
                        displayName
                        warning
                    }
                }
            `,
            { input }
        )
        .toPromise()
}
