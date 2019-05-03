import { DOMFunctions, PositionAdjuster } from '@sourcegraph/codeintellify'
import { Selection } from '@sourcegraph/extension-api-types'
import { Observable, of, zip } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { Omit } from 'utility-types'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../../shared/src/util/url'
import { ERPRIVATEREPOPUBLICSOURCEGRAPHCOM, isErrorLike } from '../../shared/backend/errors'
import { ButtonProps } from '../../shared/components/CodeViewToolbar'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { CodeHost, FileInfo } from './code_intelligence'
import { ensureRevisionsAreCloned } from './util/file_info'
import { trackViews, ViewResolver } from './views'

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
    resolveFileInfo: (codeView: HTMLElement) => Observable<FileInfo>
    /**
     * In some situations, we need to be able to adjust the position going into
     * and coming out of codeintellify. For example, Phabricator converts tabs
     * to spaces in it's DOM.
     */
    adjustPosition?: PositionAdjuster<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec>
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
export const trackCodeViews = ({ codeViewResolvers }: Pick<CodeHost, 'codeViewResolvers'>) =>
    trackViews<CodeView>(codeViewResolvers)

export interface FileInfoWithContents extends FileInfo {
    content?: string
    baseContent?: string
    headHasFileContents?: boolean
    baseHasFileContents?: boolean
}

export const fetchFileContents = (info: FileInfo): Observable<FileInfoWithContents> =>
    ensureRevisionsAreCloned(info).pipe(
        switchMap(info => {
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
