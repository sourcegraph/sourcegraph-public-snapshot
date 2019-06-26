import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { mutateGraphQL } from '../../../backend/graphql'
import { createThread, discussionThreadTargetFieldsFragment } from '../../../discussions/backend'
import { fetchRepository } from '../../../repo/settings/backend'
import { computeDiff, FileDiff } from '../../threads/detail/changes/computeDiff'
import { ChangesetDelta, ThreadSettings } from '../../threads/settings'

export const FAKE_PROJECT_ID = 'UHJvamVjdDox' // TODO!(sqs)

/**
 * The initial status for a changeset thread when creating it. {@link GQL.ThreadStatus.OPEN_ACTIVE}
 * is for "Create changeset" and {@link GQL.ThreadStatus.PREVIEW} is for "Preview changeset".
 */
export type ChangesetCreationStatus = GQL.ThreadStatus.OPEN_ACTIVE | GQL.ThreadStatus.PREVIEW

interface ChangesetCreationInfo
    extends Pick<GQL.ICreateThreadOnDiscussionsMutationArguments['input'], 'title' | 'contents'>,
        Pick<ThreadSettings, 'changesetActionDescriptions'> {
    status: ChangesetCreationStatus
}

/**
 * Create a changeset by applying the {@link codeAction}.
 */
export async function createChangesetFromCodeAction(
    { extensionsController }: ExtensionsControllerProps,
    diagnostic: sourcegraph.Diagnostic,
    codeAction: sourcegraph.CodeAction,
    info: Pick<ChangesetCreationInfo, 'status'>
): Promise<Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'url' | 'status'>> {
    return createChangesetFromDiffs(await computeDiff(extensionsController, [codeAction]), {
        ...info,
        title: `${diagnostic.message}: ${codeAction.title}`,
        contents: '',
        changesetActionDescriptions: [
            { user: 'sqs', timestamp: Date.now(), title: codeAction.title, detail: diagnostic.message },
        ],
    })
}

/**
 * Create a changeset by applying the diffs.
 */
export async function createChangesetFromDiffs(
    fileDiffs: FileDiff[],
    info: ChangesetCreationInfo
): Promise<Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'url' | 'status'>> {
    const fileDiffsByRepo = new Map<string, FileDiff[]>()
    for (const fileDiff of fileDiffs) {
        const repo = parseRepoURI(fileDiff.newPath!).repoName
        const repoFileDiffs = fileDiffsByRepo.get(repo) || []
        repoFileDiffs.push(fileDiff)
        fileDiffsByRepo.set(repo, repoFileDiffs)
    }

    const deltas: ChangesetDelta[] = []
    for (const [repoName, fileDiffs] of fileDiffsByRepo) {
        const repo = await fetchRepository(repoName).toPromise()
        const delta: ChangesetDelta = {
            repository: repo.id,
            base: 'master' /* TODO!(sqs) */,
            head: 'refs/changesets/preview/1',
        }

        const baseCommit = parseRepoURI(fileDiffs[0].newPath!).commitID!
        await gitCreateRefFromPatch({
            input: {
                repository: delta.repository,
                name: delta.head,
                baseCommit,
                patch: fileDiffs.map(({ patch }) => patch).join('\n'),
                commitMessage:
                    info.changesetActionDescriptions && info.changesetActionDescriptions.length > 0
                        ? `${info.changesetActionDescriptions
                              .map(c => c.title)
                              .join(', ')}\n\n${info.changesetActionDescriptions
                              .filter(c => c.detail)
                              .map(
                                  c =>
                                      `- ${info.changesetActionDescriptions!.length > 1 ? `${c.title}: ` : ''}${
                                          c.detail
                                      }`
                              )
                              .join('\n')}`
                        : 'Changeset commit',
            },
        }).toPromise()
        deltas.push(delta)
    }

    const settings: ThreadSettings = { deltas, changesetActionDescriptions: info.changesetActionDescriptions }
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
