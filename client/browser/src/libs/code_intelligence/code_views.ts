import { propertyIsDefined } from '@sourcegraph/codeintellify/lib/helpers'
import { Observable, of, zip } from 'rxjs'
import { catchError, concatAll, map, mergeMap } from 'rxjs/operators'

import { fetchBlobContentLines } from '../../shared/repo/backend'
import { MutationRecordLike } from '../../shared/util/dom'
import { CodeHost, FileInfo, ResolvedCodeView } from './code_intelligence'

const findCodeViews = ({ codeViewSpecs, codeViewSpecResolver }: CodeHost, nodes: Iterable<Node>): ResolvedCodeView[] =>
    [...nodes]
        .filter((node): node is HTMLElement => node instanceof HTMLElement)
        .flatMap(element => [
            ...(codeViewSpecs
                ? codeViewSpecs
                      // Find all new code views within the added element
                      // (MutationObservers don't emit all descendant nodes of an addded node recursively)
                      .map(({ selector, ...info }) => ({
                          info,
                          matches: element.matches(selector)
                              ? [element]
                              : [...element.querySelectorAll<HTMLElement>(selector)],
                      }))
                      .flatMap(({ info, matches }) => matches.map(codeViewElement => ({ ...info, codeViewElement })))
                : []),
            // code views from resolver
            ...(codeViewSpecResolver
                ? (element.matches(codeViewSpecResolver.selector)
                      ? [element]
                      : [...element.querySelectorAll<HTMLElement>(codeViewSpecResolver.selector)]
                  )
                      .map(codeViewElement => ({
                          resolved: codeViewSpecResolver.resolveCodeViewSpec(codeViewElement),
                          codeViewElement,
                      }))
                      .filter(propertyIsDefined('resolved'))
                      .map(({ resolved, ...rest }) => ({ ...resolved, ...rest }))
                : []),
        ])

/**
 * Find all the code views on a page given a CodeHostSpec from DOM mutations.
 *
 * Emits every code view that gets added or removed.
 *
 * At any given time, there can be 0-n code views on a page.
 */
export const trackCodeViews = (codeHost: CodeHost) => (
    mutations: Observable<MutationRecordLike[]>
): Observable<ResolvedCodeView & { type: 'added' | 'removed' }> =>
    mutations.pipe(
        concatAll(),
        mergeMap(mutation => [
            ...findCodeViews(codeHost, mutation.addedNodes).map(codeView => ({
                type: 'added' as 'added',
                ...codeView,
            })),
            ...findCodeViews(codeHost, mutation.removedNodes).map(codeView => ({
                type: 'removed' as 'removed',
                ...codeView,
            })),
        ])
    )
// .pipe(emitWhenIntersecting(250))

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
