import { parsePatch } from 'diff'
import H from 'history'
import { of, Subscription, throwError, Unsubscribable } from 'rxjs'
import { first, switchMap } from 'rxjs/operators'
import { parseContributionExpressions } from '../../../../../shared/src/api/client/services/contribution'
import { ContributableMenu } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { createThread } from '../../../discussions/backend'
import { parseSearchURLQuery } from '../../../search'
import { search } from '../../../search/backend'
import { FAKE_PROJECT_ID } from '../../changesets/preview/backend'
import { FileDiff, npmDiffToFileDiffHunk } from '../../threads/detail/changes/computeDiff'
import { ThreadSettings } from '../../threads/settings'
import { queryWithReplacementText } from '../query'

export const CODEMOD_PANEL_VIEW_ID = 'codemod'

export function registerCodemodSearchContributions({
    history,
    extensionsController,
}: {
    history: H.History
} & ExtensionsControllerProps<'services'>): Unsubscribable {
    const subscriptions = new Subscription()

    const REPLACE_ID = 'codemod.search.replace'
    subscriptions.add(
        extensionsController.services.commands.registerCommand({
            command: REPLACE_ID,
            run: async () => {
                const text = prompt('Enter replacement text:')
                if (text !== null) {
                    const params = new URLSearchParams(history.location.search)
                    params.set('q', queryWithReplacementText(params.get('q') || '', text))
                    history.push({
                        // TODO!(sqs):why is this commented out/necessary?
                        //
                        // ...TabsWithURLViewStatePersistence.urlForTabID(history.location, CODEMOD_PANEL_VIEW_ID),
                        search: `${params}`,
                    })
                }
            },
        })
    )
    subscriptions.add(
        extensionsController.services.contribution.registerContributions({
            contributions: parseContributionExpressions({
                actions: [
                    {
                        id: REPLACE_ID,
                        title: 'Replace',
                        category: 'Codemod',
                        command: REPLACE_ID,
                        actionItem: {
                            label: 'Replace...',
                            // TODO!(sqs): icon theme color doesn't update
                            iconURL:
                                "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' style='width:24px;height:24px' viewBox='0 0 24 24'%3E%3Cpath fill='%23a2b0cd' d='M11,6C12.38,6 13.63,6.56 14.54,7.46L12,10H18V4L15.95,6.05C14.68,4.78 12.93,4 11,4C7.47,4 4.57,6.61 4.08,10H6.1C6.56,7.72 8.58,6 11,6M16.64,15.14C17.3,14.24 17.76,13.17 17.92,12H15.9C15.44,14.28 13.42,16 11,16C9.62,16 8.37,15.44 7.46,14.54L10,12H4V18L6.05,15.95C7.32,17.22 9.07,18 11,18C12.55,18 14,17.5 15.14,16.64L20,21.5L21.5,20L16.64,15.14Z' /%3E%3C/svg%3E",
                        },
                    },
                ],
                menus: {
                    [ContributableMenu.SearchResultsToolbar]: [{ action: REPLACE_ID }],
                },
            }),
        })
    )

    const SAVE_ID = 'codemod.search.saveAsCheck'
    subscriptions.add(
        extensionsController.services.commands.registerCommand({
            command: SAVE_ID,
            run: async () => {
                // const title = prompt('Enter title to create changeset:')
                // if (title !== null) {
                const query = parseSearchURLQuery(history.location.search) || ''

                // TODO!(sqs) hack get diffs from search
                const results = await search(query, { extensionsController })
                    .pipe(
                        first(),
                        switchMap(r => {
                            if (isErrorLike(r)) {
                                // tslint:disable-next-line: rxjs-throw-error
                                return throwError(r)
                            }
                            return of(r)
                        })
                    )
                    .toPromise()
                const diffs: FileDiff[] = results.results
                    .filter((r): r is GQL.ICodemodResult => r.__typename === 'CodemodResult')
                    .flatMap(r =>
                        r.matches.flatMap(m =>
                            parsePatch(
                                `--- ${m.url}\n+++ ${m.url}\n${m.body.text
                                    .replace(/^```diff\n/, '')
                                    .replace(/\n```$/, '')
                                    .replace(/\n /g, '\n')}`
                            )
                        )
                    )
                    .map(
                        diff =>
                            ({
                                oldPath: diff.oldFileName,
                                newPath: diff.newFileName,
                                hunks: diff.hunks.map(npmDiffToFileDiffHunk),
                            } as FileDiff)
                    )

                const settings: ThreadSettings = {
                    queries: [query],
                    previewChangesetDiff: diffs,
                }
                const thread = await createThread({
                    title: query,
                    settings: JSON.stringify(settings),
                    contents: `Created from search:\n\n${'```'}\n${query}\n${'```'}`,
                    project: FAKE_PROJECT_ID,
                    status: GQL.ThreadStatus.PREVIEW,
                    type: GQL.ThreadType.CHANGESET,
                }).toPromise()
                history.push(thread.url)
                // }
            },
        })
    )
    subscriptions.add(
        extensionsController.services.contribution.registerContributions({
            contributions: parseContributionExpressions({
                actions: [
                    {
                        id: SAVE_ID,
                        title: 'Create from query',
                        category: 'Codemod',
                        command: SAVE_ID,
                        actionItem: {
                            label: 'Preview changeset',
                            // TODO!(sqs): icon theme color doesn't update
                            iconURL:
                                "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' style='width:24px;height:24px' viewBox='0 0 24 24' fill='transparent'%3E%3Cpath d='M20,16V10H22V16C22,17.1 21.1,18 20,18H8C6.89,18 6,17.1 6,16V4C6,2.89 6.89,2 8,2H16V4H8V16H20M10.91,7.08L14,10.17L20.59,3.58L22,5L14,13L9.5,8.5L10.91,7.08M16,20V22H4C2.9,22 2,21.1 2,20V7H4V20H16Z'%3E%3C/path%3E%3C/svg%3E",
                        },
                    },
                ],
                menus: {
                    [ContributableMenu.SearchResultsToolbar]: [{ action: SAVE_ID }],
                },
            }),
        })
    )

    return subscriptions
}
