import { Range } from '@sourcegraph/extension-api-classes'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import { sortBy, uniq } from 'lodash'
import { combineLatest, from, Observable, of } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { DiagnosticWithType } from '../../../../../shared/src/api/client/services/diagnosticService'
import { match } from '../../../../../shared/src/api/client/types/textDocument'
import { Action, fromAction } from '../../../../../shared/src/api/types/action'
import { toDiagnostic } from '../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { getModeFromPath } from '../../../../../shared/src/languages'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { memoizeObservable } from '../../../../../shared/src/util/memoizeObservable'
import { isDefined } from '../../../../../shared/src/util/types'
import { makeRepoURI, parseRepoURI } from '../../../../../shared/src/util/url'
import { queryGraphQL } from '../../../backend/graphql'
import { PullRequestFields, ThreadSettings } from '../settings'
import { computeDiff, FileDiff } from './changes/computeDiff'

export interface DiagnosticInfo extends DiagnosticWithType {
    entry: Pick<GQL.ITreeEntry, 'path' | 'isDirectory' | 'url'> & {
        commit: Pick<GQL.IGitCommit, 'oid'>
        repository: Pick<GQL.IRepository, 'name'>
    } & (Pick<GQL.IGitBlob, '__typename' | 'content'> | Pick<GQL.IGitTree, '__typename'>)
}

// TODO!(sqs): use relative path/rev for DiscussionThreadTargetRepo
const queryCandidateFile = memoizeObservable(
    (uri: string): Observable<[URL, DiagnosticInfo['entry']]> => {
        const parsed = parseRepoURI(uri)
        return queryGraphQL(
            gql`
                query CandidateFile($repo: String!, $rev: String!, $path: String!) {
                    repository(name: $repo) {
                        commit(rev: $rev) {
                            blob(path: $path) {
                                path
                                content
                                repository {
                                    name
                                }
                                commit {
                                    oid
                                }
                                url
                            }
                        }
                    }
                }
            `,
            { repo: parsed.repoName, rev: parsed.rev || parsed.commitID, path: parsed.filePath }
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.blob ||
                    (errors && errors.length > 0)
                ) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.blob
            }),
            map(data => [new URL(uri), data] as [URL, DiagnosticInfo['entry']])
        )
    },
    uri => uri.toString()
)

export const queryCandidateFiles = async (uris: string[]): Promise<[URL, DiagnosticInfo['entry']][]> =>
    Promise.all(uris.map(uri => queryCandidateFile(uri).toPromise()))

export const toDiagnosticInfos = async (diagnostics: DiagnosticWithType[]) => {
    const uniqueResources = uniq(diagnostics.map(d => d.resource.toString()))
    const entries = await queryCandidateFiles(uniqueResources)
    const m = new Map<string, DiagnosticInfo['entry']>()
    for (const [url, entry] of entries) {
        m.set(url.toString(), entry)
    }
    return diagnostics.map(diag => {
        const entry = m.get(diag.resource.toString())
        if (!entry) {
            throw new Error(`no entry for url ${diag.resource}`)
        }
        const info: DiagnosticInfo = { ...diag, entry }
        return info
    })
}

/**
 * @param query Only observe diagnostics matching the {@link sourcegraph.DiagnosticQuery}.
 */
export const getDiagnosticInfos = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    query?: sourcegraph.DiagnosticQuery
): Observable<DiagnosticInfo[]> =>
    from(
        extensionsController.services.diagnostics
            .observeDiagnostics({}, query ? query.type : undefined)
            .pipe(map(diagnostics => (query ? diagnostics.filter(diagnosticQueryMatcher(query)) : diagnostics)))
    ).pipe(switchMap(diagEntries => toDiagnosticInfos(diagEntries)))

export function diagnosticQueryMatcher(
    query: sourcegraph.DiagnosticQuery
): (diagnostic: DiagnosticWithType) => boolean {
    return diagnostic =>
        diagnostic.type === query.type &&
        (!query.document ||
            match([query.document], {
                uri: diagnostic.resource.toString(),
                languageId: /*TODO!(sqs)*/ getModeFromPath(diagnostic.resource.pathname),
            })) &&
        (!query.range || query.range.isEqual(diagnostic.range)) &&
        (query.tag !== undefined ? !!diagnostic.tags && diagnostic.tags.includes(query.tag) : true)
}

// TODO!(sqs): this assumes a diag provider never has 2 diagnostics on the same range and resource
export const diagnosticID = (diagnostic: DiagnosticWithType): string =>
    `${diagnostic.type}:${diagnostic.resource.toString()}:${diagnostic.range ? diagnostic.range.start.line : '-'}:${
        diagnostic.range ? diagnostic.range.start.character : '-'
    }`

// TODO!(sqS): this is a bad idea because there is no canonical json representation of diagnosticquery
export const diagnosticQueryKey = (query: sourcegraph.DiagnosticQuery): string =>
    JSON.stringify({ ...query, range: query.range ? (query.range as any).toJSON() : undefined })

export const diagnosticQueryForSingleDiagnostic = (diagnostic: DiagnosticWithType): sourcegraph.DiagnosticQuery => {
    if (diagnostic.tags && diagnostic.tags.length >= 2) {
        throw new Error(
            'TODO!(sqs) not supported because DiagnosticQuery#tag is singleton for simplicity now, but that can be easily improved if needed'
        )
    }
    return {
        type: diagnostic.type,
        document: { pattern: diagnostic.resource.toString() },
        range: diagnostic.range,
        tag: diagnostic.tags ? diagnostic.tags[0] : undefined,
    }
}

export const getCodeActions = memoizeObservable(
    ({
        diagnostic,
        extensionsController,
    }: { diagnostic: DiagnosticInfo } & ExtensionsControllerProps): Observable<Action[]> =>
        from(
            extensionsController.services.codeActions.getCodeActions({
                textDocument: {
                    uri: makeRepoURI({
                        repoName: diagnostic.entry.repository.name,
                        rev: diagnostic.entry.commit.oid,
                        commitID: diagnostic.entry.commit.oid,
                        filePath: diagnostic.entry.path,
                    }),
                },
                range: Range.fromPlain(diagnostic.range),
                context: { diagnostics: [diagnostic] },
            })
        ).pipe(
            map(codeActions => codeActions || []),
            map(actions => actions.map(fromAction))
        ),
    ({ diagnostic }) => diagnosticID(diagnostic)
)

export const codeActionID = (codeAction: Action): string => codeAction.title // TODO!(sqs): codeAction.title is not guaranteed unique

export const getActiveCodeAction0 = (
    diagnostic: DiagnosticInfo,
    threadSettings: ThreadSettings,
    codeActions: Action[]
): Action | undefined => {
    const activeCodeActionID =
        threadSettings && threadSettings.actions && threadSettings.actions[diagnosticID(diagnostic)]
    return codeActions.find(a => codeActionID(a) === activeCodeActionID) || codeActions[0]
}

export const getActiveCodeAction = (
    diagnostic: DiagnosticInfo,
    extensionsController: ExtensionsControllerProps['extensionsController'],
    threadSettings: ThreadSettings
): Observable<Action | undefined> =>
    getCodeActions({ diagnostic, extensionsController }).pipe(
        map(codeActions => getActiveCodeAction0(diagnostic, threadSettings, codeActions))
    )

export interface Changeset {
    thread: Pick<GQL.IDiscussionThread, 'id'>
    repo: string
    pullRequest: PullRequestFields
    fileDiffs: FileDiff[]
}

const interpolatePullRequestTemplate = ({ title, branch, description }: PullRequestFields): PullRequestFields => ({
    title,
    branch,
    description: description
        .replace('${check_number}', '49')
        .replace('${check_url}', 'https://sourcegraph.example.com/checks/49')
        .replace(
            '${related_links}',
            '- [sourcegraph/codeintellify#41](#)\n- [sourcegraph/sourcegraph#9184](#)\n- [sourcegraph/react-loading-spinner#35](#)'
        ),
})

export const computeChangesets = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    thread: Pick<GQL.IDiscussionThread, 'id'>,
    threadSettings: ThreadSettings,
    query?: { repo: string }
): Observable<Changeset[]> =>
    getDiagnosticInfos(extensionsController).pipe(
        map(diagnostics => (query ? diagnostics.filter(d => d.entry.repository.name === query.repo) : diagnostics)),
        switchMap(diagnostics =>
            diagnostics.length > 0
                ? combineLatest(diagnostics.map(d => getActiveCodeAction(d, extensionsController, threadSettings)))
                : of([])
        ),
        switchMap(codeActions => computeDiff(extensionsController, codeActions.filter(isDefined))),
        map(fileDiffs => {
            const byRepo = new Map<string, FileDiff[]>()
            for (const fileDiff of fileDiffs) {
                const parsed = parseRepoURI(fileDiff.newPath || fileDiff.oldPath!)
                const key = parsed.repoName
                byRepo.set(key, [...(byRepo.get(key) || []), fileDiff])
            }

            const changesets: Changeset[] = []
            for (const [repo, fileDiffs] of byRepo) {
                changesets.push({
                    thread,
                    repo,
                    pullRequest: interpolatePullRequestTemplate({
                        title: 'Untitled',
                        branch: 'codemod-84571', // TODO!(sqs)
                        description: 'No description set',
                        ...threadSettings.pullRequestTemplate,
                    }),
                    fileDiffs,
                })
            }
            return sortBy(changesets, c => c.repo)
        })
    )

export type ChangesetExternalStatus = 'open' | 'merged' | 'closed'

const CHANGESET_EXTERNAL_STATUSES: ChangesetExternalStatus[] = ['open', 'merged', 'closed']

export const getChangesetExternalStatus = ({
    repo,
    fileDiffs,
    thread,
}: Pick<Changeset, 'repo' | 'thread'> & { fileDiffs: { length: number } }): {
    title: string
    status: ChangesetExternalStatus
    commentsCount: number
} => {
    const k = (repo + thread.id).split('').reduce((sum, c) => (sum += c.charCodeAt(0)), 0) + fileDiffs.length
    const status = CHANGESET_EXTERNAL_STATUSES[k % CHANGESET_EXTERNAL_STATUSES.length]
    return {
        title: `#${k % 300}`,
        status: status === 'closed' && k % 20 === 0 ? 'closed' : k % 2 === 0 ? 'merged' : 'open',
        commentsCount: k % 17,
    }
}
