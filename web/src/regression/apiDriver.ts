/**
 * Provides convenience functions for interacting with the Sourcegraph API from tests.
 */

import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { GraphQLClient } from './api'

export class APIDriver {
    constructor(public gqlClient: GraphQLClient) {}

    /**
     * Block until all repositories in the list exist
     */
    public async waitForRepos(ensureRepos: string[]): Promise<void> {
        for (const repoName of ensureRepos) {
            while (true) {
                const res = await this.gqlClient
                    .queryGraphQL(
                        gql`
                            query ResolveRev($repoName: String!, $rev: String!) {
                                repository(name: $repoName) {
                                    mirrorInfo {
                                        cloneInProgress
                                        cloneProgress
                                        cloned
                                    }
                                    commit(rev: $rev) {
                                        oid
                                        tree(path: "") {
                                            url
                                        }
                                    }
                                    defaultBranch {
                                        abbrevName
                                    }
                                    redirectURL
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
                    break
                }
                // Wait 1s
                console.log(`Repository ${repoName} did not exist, waiting 1s before polling again.`)
                await new Promise(resolve => setTimeout(() => resolve(), 1000))
            }
        }
    }

    public async ensureExternalService(options: {
        kind: GQL.ExternalServiceKind
        uniqueDisplayName: string
        config: Record<string, any>
    }): Promise<void> {
        const externalServicesRes = await this.gqlClient
            .queryGraphQL(
                gql`
                    query ExternalServices($first: Int) {
                        externalServices(first: $first) {
                            nodes {
                                id
                                kind
                                displayName
                                config
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
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
        await this.gqlClient
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
}
