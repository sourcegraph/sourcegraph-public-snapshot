import { from, merge, Observable, of, zip } from 'rxjs'
import { catchError, concatAll, filter, map, mergeMap } from 'rxjs/operators'
import { isDefined, isInstanceOf } from '../../../../../shared/src/util/types'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { MutationRecordLike, querySelectorAllOrSelf } from '../../shared/util/dom'
import { DiffViewSpecResolver, FileInfo, ResolvedDiffView } from './code_intelligence'

export interface AddedDiffView extends ResolvedDiffView {
    type: 'added'
}

export interface RemovedDiffView extends Pick<ResolvedDiffView, 'diffViewElement'> {
    type: 'removed'
}

export type DiffViewEvent = AddedDiffView | RemovedDiffView

/**
 * Find all the code views on a page given a CodeHostSpec from DOM mutations.
 *
 * Emits every code view that gets added or removed.
 *
 * At any given time, there can be 0-n code views on a page.
 */
export const trackDiffViews = (diffViewSpecResolvers: DiffViewSpecResolver[]) => (
    mutations: Observable<MutationRecordLike[]>
): Observable<DiffViewEvent> =>
    mutations.pipe(
        concatAll(),
        mergeMap(mutation =>
            merge(
                // Find all new code views within the added nodes
                // (MutationObservers don't emit all descendant nodes of an addded node recursively)
                from(mutation.addedNodes).pipe(
                    filter(isInstanceOf(HTMLElement)),
                    mergeMap(addedElement =>
                        from(diffViewSpecResolvers).pipe(
                            mergeMap(spec =>
                                [...(querySelectorAllOrSelf(addedElement, spec.selector) as Iterable<HTMLElement>)]
                                    .map(diffViewElement => {
                                        const diffViewSpec = spec.resolveDiffViewSpec(diffViewElement)
                                        return (
                                            diffViewSpec && {
                                                ...diffViewSpec,
                                                diffViewElement,
                                                type: 'added' as const,
                                            }
                                        )
                                    })
                                    .filter(isDefined)
                            )
                        )
                    )
                ),
                // For removed nodes, find the removed elements, but don't resolve the kind (it's not relevant)
                from(mutation.removedNodes).pipe(
                    filter(isInstanceOf(HTMLElement)),
                    mergeMap(removedElement =>
                        from(diffViewSpecResolvers).pipe(
                            mergeMap(
                                ({ selector }) =>
                                    querySelectorAllOrSelf(removedElement, selector) as Iterable<HTMLElement>
                            ),
                            map(diffViewElement => ({
                                diffViewElement,
                                type: 'removed' as const,
                            }))
                        )
                    )
                )
            )
        )
    )

export interface FileInfoWithContents extends FileInfo {
    content?: string
    baseContent?: string
    headHasFileContents?: boolean
    baseHasFileContents?: boolean
}

export const fetchFileContents = (info: FileInfo): Observable<FileInfoWithContents> => {
    const fetchingBaseFile = info.baseCommitID
        ? fetchBlobContentLines({
              repoName: info.repoName,
              filePath: info.baseFilePath || info.filePath,
              commitID: info.baseCommitID,
          })
        : of(null)

    const fetchingHeadFile = fetchBlobContentLines({
        repoName: info.repoName,
        filePath: info.filePath,
        commitID: info.commitID,
    })

    return zip(fetchingBaseFile, fetchingHeadFile).pipe(
        map(
            ([baseFileContent, headFileContent]): FileInfoWithContents => ({
                ...info,
                baseContent: baseFileContent ? baseFileContent.join('\n') : undefined,
                content: headFileContent.join('\n'),
                headHasFileContents: headFileContent.length > 0,
                baseHasFileContents: baseFileContent ? baseFileContent.length > 0 : undefined,
            })
        ),
        catchError(error => [info])
    )
}
