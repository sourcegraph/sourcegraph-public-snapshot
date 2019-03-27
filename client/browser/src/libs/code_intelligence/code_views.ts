import { from, merge, Observable, of, zip } from 'rxjs'
import { catchError, concatAll, filter, map, mergeMap } from 'rxjs/operators'
import { isDefined } from '../../../../../shared/src/util/types'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { MutationRecordLike, querySelectorAllOrSelf } from '../../shared/util/dom'
import { CodeHost, CodeViewSpec, CodeViewSpecResolver, FileInfo, ResolvedCodeView } from './code_intelligence'

export interface AddedCodeView extends ResolvedCodeView {
    type: 'added'
}

export interface RemovedCodeView extends Pick<ResolvedCodeView, 'codeViewElement'> {
    type: 'removed'
}

export type CodeViewEvent = AddedCodeView | RemovedCodeView

const isHTMLElement = (node: unknown): node is HTMLElement => node instanceof HTMLElement

/** Converts a static CodeViewSpec to a dynamic CodeViewSpecResolver */
const toCodeViewResolver = ({ selector, ...spec }: CodeViewSpec): CodeViewSpecResolver => ({
    selector,
    resolveCodeViewSpec: () => spec,
})

/**
 * Find all the code views on a page given a CodeHostSpec from DOM mutations.
 *
 * Emits every code view that gets added or removed.
 *
 * At any given time, there can be 0-n code views on a page.
 */
export const trackCodeViews = ({
    codeViewSpecs = [],
    codeViewSpecResolver,
}: Pick<CodeHost, 'codeViewSpecs' | 'codeViewSpecResolver'>) => (
    mutations: Observable<MutationRecordLike[]>
): Observable<CodeViewEvent> => {
    const codeViewSpecResolvers = codeViewSpecs.map(toCodeViewResolver)
    if (codeViewSpecResolver) {
        codeViewSpecResolvers.push(codeViewSpecResolver)
    }
    return mutations.pipe(
        concatAll(),
        mergeMap(mutation =>
            merge(
                // Find all new code views within the added nodes
                // (MutationObservers don't emit all descendant nodes of an addded node recursively)
                from(mutation.addedNodes).pipe(
                    filter(isHTMLElement),
                    mergeMap(addedElement =>
                        from(codeViewSpecResolvers).pipe(
                            mergeMap(spec =>
                                [...(querySelectorAllOrSelf(addedElement, spec.selector) as Iterable<HTMLElement>)]
                                    .map(codeViewElement => {
                                        const codeViewSpec = spec.resolveCodeViewSpec(codeViewElement)
                                        return (
                                            codeViewSpec && {
                                                ...codeViewSpec,
                                                codeViewElement,
                                                type: 'added' as 'added',
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
                    filter(isHTMLElement),
                    mergeMap(removedElement =>
                        from(codeViewSpecResolvers).pipe(
                            mergeMap(
                                ({ selector }) =>
                                    querySelectorAllOrSelf(removedElement, selector) as Iterable<HTMLElement>
                            ),
                            map(codeViewElement => ({
                                codeViewElement,
                                type: 'removed' as 'removed',
                            }))
                        )
                    )
                )
            )
        )
    )
}

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
