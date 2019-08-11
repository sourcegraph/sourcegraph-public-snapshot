import { Observable } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { authenticatedUser } from '../../../auth'
import { mutateGraphQL, queryGraphQL } from '../../../backend/graphql'
import { fetchRepository } from '../../../repo/settings/backend'
import { createThread } from '../../threads/repository/new/ThreadsNewPage'
import { FileDiff } from '../../threadsOLD/detail/changes/computeDiff'
import { GitHubPRLink } from '../../threadsOLD/settings'
import { addThreadsToCampaign } from '../detail/threads/AddThreadToCampaignDropdownButton'
import { createCampaign } from '../namespace/new/CampaignsNewPage'

export const FAKE_PROJECT_ID = 'UHJvamVjdDox' // TODO!(sqs)

interface CampaignCreationInfo extends Pick<GQL.ICreateCampaignInput, 'name' | 'description' | 'preview' | 'rules'> {}

/**
 * Create a campaign by applying the diffs.
 */
export async function createCampaignFromDiffs(
    fileDiffs: FileDiff[],
    info: CampaignCreationInfo
): Promise<Pick<GQL.ICampaign, 'id' | 'url'>> {
    const githubToken = localStorage.getItem('githubToken')
    if (githubToken === null) {
        throw new Error(
            "You must set a GitHub access token (`localStorage.githubToken='...'` in your browser's JavaScript console) then try again."
        )
    }

    const authedUser = await authenticatedUser.pipe(first()).toPromise()
    if (!authedUser) {
        throw new Error('not signed in')
    }
    const campaign = await createCampaign({
        ...info,
        namespace: authedUser.id,
    })

    const fileDiffsByRepo = new Map<string, FileDiff[]>()
    for (const fileDiff of fileDiffs) {
        const repo = parseRepoURI(fileDiff.newPath!).repoName
        const repoFileDiffs = fileDiffsByRepo.get(repo) || []
        repoFileDiffs.push(fileDiff)
        fileDiffsByRepo.set(repo, repoFileDiffs)
    }

    const changesets: Promise<GQL.IThread>[] = []
    for (const [repoName, fileDiffs] of fileDiffsByRepo) {
        const repo = (await queryGraphQL(
            gql`
                query RepositoryForChangeset($name: String!) {
                    repository(name: $name) {
                        id
                        commit(rev: "HEAD") {
                            oid
                        }
                    }
                }
            `,
            { name: repoName }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => data.repository)
            )
            .toPromise())!

        const baseRef = 'refs/heads/master' /* TODO!(sqs) */
        const branchName = info.name.replace(/[^a-zA-Z0-9]+/g, '-').toLowerCase() // TODO!(sqs)
        const headRef = `refs/heads/${branchName}` /* TODO!(sqs) */

        const baseCommit = repo.commit!.oid

        // TODO!(sqs) hack create github PRs
        const title = info.name || 'Sourcegraph changeset'
        changesets.push(
            gitCreateRefFromPatch({
                input: {
                    repository: repo.id,
                    name: headRef,
                    baseCommit,
                    patch: fileDiffs.map(({ patch }) => patch).join('\n'),
                    commitMessage: info.name,
                },
            })
                .toPromise()
                .then(() =>
                    createGitHubPR(githubToken, {
                        repositoryName: repoName,
                        baseRefName: baseRef,
                        headRefName: headRef,
                        title,
                    })
                )
                .then(prLink =>
                    createThread({
                        repository: repo.id,
                        title,
                        preview: info.preview,
                        baseRef,
                        headRef,
                        externalURL: prLink.url, // TODO!(sqs)
                    })
                )
        )
    }

    await Promise.all(changesets).then(changesets =>
        addThreadsToCampaign({ campaign: campaign.id, threads: changesets.map(({ id }) => id) })
    )
    return campaign
}

function gitCreateRefFromPatch(
    args: GQL.ICreateRefFromPatchOnGitMutationArguments
): Observable<GQL.IGitCreateRefFromPatchPayload> {
    return mutateGraphQL(
        gql`
            mutation CreateRefFromPatch($input: GitCreateRefFromPatchInput!) {
                git {
                    createRefFromPatch(input: $input) {
                        ref {
                            name
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.git || !data.git.createRefFromPatch || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.git.createRefFromPatch
        })
    )
}

async function createGitHubPR(
    githubToken: string,
    {
        repositoryName,
        baseRefName,
        headRefName,
        title,
    }: {
        repositoryName: string
        baseRefName: string
        headRefName: string
        title: string
    }
): Promise<GitHubPRLink> {
    const GITHUB_GRAPHQL_API_URL = 'https://api.github.com/graphql'
    const githubRequestGraphQL = async (query: string, variables: { [name: string]: any }) => {
        const resp = await fetch(GITHUB_GRAPHQL_API_URL, {
            method: 'POST',
            headers: { Authorization: `bearer ${githubToken}`, 'Content-Type': 'application/json; charset=utf-8' },
            body: JSON.stringify({ query, variables }),
        })
        if (resp.status !== 200) {
            throw new Error(`Error from ${GITHUB_GRAPHQL_API_URL}: ${await resp.text()}`)
        }
        const { data, errors } = await resp.json()
        if (errors && errors.length > 0) {
            throw createAggregateError(errors)
        }
        return data
    }

    const [, owner, name] = repositoryName.split('/')
    const {
        repository: { id: repositoryId },
    } = await githubRequestGraphQL(
        `
    query ($owner: String!, $name: String!) {
        repository(owner: $owner, name: $name) {
            id
        }
    }`,
        { owner, name }
    )

    interface CreatePullRequestPayload {
        pullRequest: { id: GQL.ID; number: number; url: string }
    }
    try {
        const payload: { createPullRequest: CreatePullRequestPayload } = await githubRequestGraphQL(
            `
    mutation ($input: CreatePullRequestInput!) {
        createPullRequest(input: $input) {
            pullRequest {
                id
                number
                url
            }
        }
    }`,
            { input: { repositoryId, baseRefName, headRefName, title } }
        )
        return { ...payload.createPullRequest.pullRequest, repositoryName }
    } catch (err) {
        if (err.message.includes('A pull request already exists')) {
            const prs: {
                node: { pullRequests: { nodes: CreatePullRequestPayload['pullRequest'][] } }
            } = await githubRequestGraphQL(
                `
            query ($repositoryId: ID!, $headRefName: String!) {
                node(id: $repositoryId) {
                    ... on Repository {
                        pullRequests(first: 1, headRefName: $headRefName)  {
                            nodes {
                                id
                                number
                                url
                            }
                        }
                    }
                }
            }`,
                { repositoryId, headRefName }
            )
            return { ...prs.node.pullRequests.nodes[0], repositoryName }
        }
        throw err
    }
}
