import { DOMFunctions, PositionAdjuster } from '@sourcegraph/codeintellify'
import { Observable, of, zip } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../../shared/src/util/url'
import { ERPRIVATEREPOPUBLICSOURCEGRAPHCOM, isErrorLike } from '../../shared/backend/errors'
import { ButtonProps } from '../../shared/components/CodeViewToolbar'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { CodeHost, FileInfo } from './code_intelligence'
import { ensureRevisionsAreCloned } from './util/file_info'
import { trackViews, ViewResolver } from './views'

/**
 * Defines a type of code view a given code host can have. It tells us how to
 * look for the code view and how to do certain things when we find it.
 */
export interface CodeViewSpec {
    /** A selector used by `document.querySelectorAll` to find the code view. */
    selector: string
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

    isDiff?: boolean
}

/**
 * Resolves a code view element found by a {@link CodeViewSpec} to a {@link ResolvedCodeView}.
 */
export interface CodeViewSpecResolver extends Pick<CodeViewSpec, Exclude<keyof CodeViewSpec, 'selector'>> {}

/**
 * A code view found on the page.
 */
export interface ResolvedCodeView extends CodeViewSpecResolver {
    /** The code view's HTML element. */
    element: HTMLElement
}

/** Converts a static CodeViewSpec to a dynamic CodeViewSpecResolver. */
const toCodeViewResolver = ({ selector, ...spec }: CodeViewSpec): ViewResolver<ResolvedCodeView> => ({
    selector,
    resolveView: element => ({ ...spec, element }),
})

/**
 * Find all the code views on a page using both the code view specs and the code view spec
 * resolvers, calling down to {@link trackViews}.
 */
export const trackCodeViews = ({
    codeViewSpecs = [],
    codeViewSpecResolver,
}: Pick<CodeHost, 'codeViewSpecs' | 'codeViewSpecResolver'>) => {
    const codeViewSpecResolvers = codeViewSpecs.map(toCodeViewResolver)
    if (codeViewSpecResolver) {
        codeViewSpecResolvers.push({
            selector: codeViewSpecResolver.selector,
            resolveView: element => {
                const view = codeViewSpecResolver.resolveView(element)
                return view ? { ...view, element } : null
            },
        })
    }
    return trackViews<ResolvedCodeView>(codeViewSpecResolvers)
}

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
