import { DiffPart, DOMFunctions as CodeIntellifyDOMFuncions, PositionAdjuster } from '@sourcegraph/codeintellify'
import { Selection } from '@sourcegraph/extension-api-types'
import { Observable, of, zip, OperatorFunction } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { Omit } from 'utility-types'
import { ERPRIVATEREPOPUBLICSOURCEGRAPHCOM } from '../../../../shared/src/backend/errors'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../shared/src/util/url'
import { ButtonProps } from '../../shared/components/CodeViewToolbar'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { CodeHost, FileInfo, FileInfoWithRepoNames } from './code_intelligence'
import { ensureRevisionsAreCloned } from './util/file_info'
import { trackViews, ViewResolver, ViewWithSubscriptions } from './views'
import { MutationRecordLike } from '../../shared/util/dom'

export interface DOMFunctions extends CodeIntellifyDOMFuncions {
    /**
     * Gets the element for the entire line. This element is used for whole-line
     * background decorations. It should span the entire width of the line
     * independent on how long the code on that line is. This may be a parent
     * element of the code element, but keep in mind that even in split diff
     * views it must only contain the line the given diff part.
     */
    getLineElementFromLineNumber: (codeView: HTMLElement, line: number, part?: DiffPart) => HTMLElement | null
}

/**
 * Defines a code view that is present on a page.
 * Exposes operations for manipulating it, and CSS classes to be applied to injected UI elements.
 */
export interface CodeView {
    /**
     * The code view element on the page.
     */
    element: HTMLElement
    /** The DOMFunctions for the code view. */
    dom: DOMFunctions
    /**
     * Finds or creates a DOM element where we should inject the
     * `CodeViewToolbar`. This function is responsible for ensuring duplicate
     * mounts aren't created.
     */
    getToolbarMount?: (codeView: HTMLElement) => HTMLElement
    /**
     * Resolves the file info for a given code view. It returns an observable
     * because some code hosts need to resolve this asynchronously. The
     * observable should only emit once.
     */
    resolveFileInfo: (codeView: HTMLElement, requestGraphQL: PlatformContext['requestGraphQL']) => Observable<FileInfo>
    /**
     * In some situations, we need to be able to adjust the position going into
     * and coming out of codeintellify. For example, Phabricator converts tabs
     * to spaces in it's DOM.
     */
    getPositionAdjuster?: (
        requestGraphQL: PlatformContext['requestGraphQL']
    ) => PositionAdjuster<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec>
    /** Props for styling the buttons in the `CodeViewToolbar`. */
    toolbarButtonProps?: ButtonProps
    /**
     * Gets the current selections for a code view.
     */
    getSelections?: (codeViewElement: HTMLElement) => Selection[]
    /**
     * Returns a stream of selections changes for a code view.
     */
    observeSelections?: (codeViewElement: HTMLElement) => Observable<Selection[]>

    /**
     * Returns the scrollBoundaries of the code view, used by codeintellify.
     * This is called once per code view, when calling Hoverifier.hoverify().
     */
    getScrollBoundaries?: (codeViewElement: HTMLElement) => HTMLElement[]
}

/**
 * Builds a CodeViewResolver from a static CodeView and a selector.
 */
export const toCodeViewResolver = (selector: string, spec: Omit<CodeView, 'element'>): ViewResolver<CodeView> => ({
    selector,
    resolveView: element => ({ ...spec, element }),
})

/**
 * Find all the code views on a page using both the code view specs and the code view spec
 * resolvers, calling down to {@link trackViews}.
 */
export const trackCodeViews = ({
    codeViewResolvers,
}: Pick<CodeHost, 'codeViewResolvers'>): OperatorFunction<MutationRecordLike[], ViewWithSubscriptions<CodeView>> =>
    trackViews<CodeView>(codeViewResolvers)

export interface FileInfoWithContents extends FileInfoWithRepoNames {
    content?: string
    baseContent?: string
    headHasFileContents?: boolean
    baseHasFileContents?: boolean
}

export const fetchFileContents = (
    info: FileInfoWithRepoNames,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfoWithContents> =>
    ensureRevisionsAreCloned(info, requestGraphQL).pipe(
        switchMap(info => {
            const fetchingBaseFile = info.baseCommitID
                ? fetchBlobContentLines({
                      repoName: info.repoName,
                      filePath: info.baseFilePath || info.filePath,
                      commitID: info.baseCommitID,
                      requestGraphQL,
                  })
                : of(null)

            const fetchingHeadFile = fetchBlobContentLines({
                repoName: info.repoName,
                filePath: info.filePath,
                commitID: info.commitID,
                requestGraphQL,
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
                catchError(() => [info])
            )
        }),
        catchError((err: any) => {
            if (isErrorLike(err) && err.code === ERPRIVATEREPOPUBLICSOURCEGRAPHCOM) {
                return [info]
            }
            throw err
        })
    )
