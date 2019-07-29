import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { mutateGraphQL } from '../../../backend/graphql'
import { createThread } from '../../../discussions/backend'
import { fetchRepository } from '../../../repo/settings/backend'
import { FileDiff } from '../../threadsOLD/detail/changes/computeDiff'
import { ChangesetDelta, GitHubPRLink, ThreadSettings } from '../../threadsOLD/settings'

export const FAKE_PROJECT_ID = 'UHJvamVjdDox' // TODO!(sqs)

/**
 * The initial status for a changeset thread when creating it. {@link GQL.ThreadStatus.OPEN}
 * is for "Create changeset" and {@link GQL.ThreadStatus.PREVIEW} is for "Preview changeset".
 */
export type ChangesetCreationStatus = GQL.ThreadStatus.OPEN | GQL.ThreadStatus.PREVIEW

interface ChangesetCreationInfo
    extends Required<Pick<GQL.ICreateThreadOnDiscussionsMutationArguments['input'], 'title' | 'contents'>>,
        Pick<ThreadSettings, 'plan'> {
    status: ChangesetCreationStatus
}

/**
 * Create a changeset by applying the diffs.
 */
export async function createChangesetFromDiffs(
    fileDiffs: FileDiff[],
    info: ChangesetCreationInfo
): Promise<Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'url' | 'status'>> {
    const githubToken = localStorage.getItem('githubToken')
    if (githubToken === null) {
        throw new Error(
            "You must set a GitHub access token (`localStorage.githubToken='...'` in your browser's JavaScript console) then try again."
        )
    }

    const fileDiffsByRepo = new Map<string, FileDiff[]>()
    for (const fileDiff of fileDiffs) {
        const repo = parseRepoURI(fileDiff.newPath!).repoName
        const repoFileDiffs = fileDiffsByRepo.get(repo) || []
        repoFileDiffs.push(fileDiff)
        fileDiffsByRepo.set(repo, repoFileDiffs)
    }

    const deltas: Promise<ChangesetDelta>[] = []
    const relatedPRs: Promise<GitHubPRLink>[] = []
    for (const [repoName, fileDiffs] of fileDiffsByRepo) {
        const repo = await fetchRepository(repoName).toPromise()
        const branchName = info.title!.replace(/[^a-zA-Z0-9]+/g, '-').toLowerCase() // TODO!(sqs)
        const delta: ChangesetDelta = {
            repository: repo.id,
            base: 'refs/heads/master' /* TODO!(sqs) */,
            head: `refs/heads/${branchName}` /* TODO!(sqs) */,
        }

        const baseCommit = parseRepoURI(fileDiffs[0].newPath!).commitID!
        deltas.push(
            gitCreateRefFromPatch({
                input: {
                    repository: delta.repository,
                    name: delta.head,
                    baseCommit,
                    patch: fileDiffs.map(({ patch }) => patch).join('\n'),
                    commitMessage:
                        info.plan && info.plan.operations.length > 0
                            ? info.plan.operations.map(c => c.message).join(', ')
                            : 'Changeset commit',
                },
            })
                .toPromise()
                .then(() => delta)
        )

        // TODO!(sqs) hack create github PRs
        relatedPRs.push(
            createGitHubPR(githubToken, {
                repositoryName: repoName,
                baseRefName: delta.base,
                headRefName: delta.head,
                title: info.title || 'Sourcegraph changeset',
            })
        )
    }

    const settings: ThreadSettings = {
        deltas: await Promise.all(deltas),
        plan: info.plan,
        relatedPRs: await Promise.all(await relatedPRs),
    }
    return createThread({
        ...info,
        type: GQL.ThreadType.CHANGESET,
        project: FAKE_PROJECT_ID,
        settings: JSON.stringify(settings, null, 2),
    }).toPromise()
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
