import * as React from 'react'
import { merge, of, Subscription } from 'rxjs'
import { catchError, delay, takeUntil } from 'rxjs/operators'
import { asError } from '../../shared/backend/errors'
import { EMODENOTFOUND, fetchJumpURL, SimpleProviderFns } from '../../shared/backend/lsp'
import { AbsoluteRepoFilePosition } from '../../shared/repo'
import { parseBrowserRepoURL } from '../../shared/repo'
import { hideTooltip, TooltipData, updateTooltip } from '../../shared/repo/tooltips'
import { eventLogger } from '../../shared/util/context'
import { BitbucketState, scrollToLine } from './utils/util'

interface Props {
    element: HTMLElement
    event?: MouseEvent
    bitbucketState: BitbucketState
    revState: {
        baseRev: string
        headRev: string
    }
    docked?: boolean
    simpleProviderFns: SimpleProviderFns
}

interface State {
    tooltipData?: TooltipData
}

const LOADING: 'loading' = 'loading'

/** The time in ms after which to show a loader if the result has not returned yet */
const LOADER_DELAY = 300

export class TooltipPortal extends React.Component<Props, State> {
    private tooltipRef: HTMLElement | null
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            tooltipData: undefined,
        }
    }

    public componentDidMount(): void {
        this.handleTooltips()
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.bitbucketState !== nextProps.bitbucketState || this.props.element !== nextProps.element) {
            this.handleTooltips()
        }
    }

    public componentWillUnmount(): void {
        hideTooltip()
        this.subscriptions.unsubscribe()
    }

    private handleTooltips(): void {
        if (!this.props.event) {
            console.warn('Err: No fromEvent.')
            return
        }
        const { srcElement } = this.props.event
        if (!srcElement) {
            console.warn('Err: No srcElement.')
            return
        }
        const pre = (srcElement as HTMLElement).closest('pre')
        if (!pre || !pre.firstChild) {
            console.warn('Err: No pre element.')
            return
        }
        const lineData = this.getLineNumberForPullRequest()
        let lineNumber: number | undefined
        if (lineData && lineData.lineNumber) {
            lineNumber = lineData.lineNumber
        } else {
            lineNumber = this.getLineNumberFromGutter(pre)
        }
        let file: string | undefined
        let repoPath: string | undefined
        let rev: string | undefined
        const { filePath, repository } = this.props.bitbucketState
        if (repository) {
            repoPath = `${window.location.hostname}/${repository.project.key}/${repository.slug}`
        }
        if (filePath) {
            // Check if we can get the header otherwise use something else.
            file = filePath.components.join('/')
        }
        // Various pages contain specific information in the toolbar. This is helpful when the activity feed shows a
        // diff across commits that are not stored in the bitbucket state object.
        const fileCommitState = this.getFileCommitStateFromToolbar()
        if (fileCommitState) {
            file = fileCommitState.file
            rev = fileCommitState.rev
        }
        if (!file) {
            const fileFromHash = window.location.hash && window.location.hash.substr(1)
            if (!fileFromHash) {
                console.error('no file present', this.props.bitbucketState)
                return
            }
            file = fileFromHash
        }

        let characterOffset = 1
        const { element } = this.props
        function getTargetOffset(root: HTMLElement): boolean | undefined {
            for (const innerSpan of root.childNodes) {
                if (innerSpan === element) {
                    return true
                }
                const content = innerSpan.textContent!
                const innerElement = innerSpan as HTMLElement
                if (!innerElement.children || innerElement.children.length === 0) {
                    if (
                        (innerSpan as HTMLElement).classList &&
                        (innerSpan as HTMLElement).classList.contains('cm-tab')
                    ) {
                        characterOffset += 1
                    } else {
                        characterOffset += content.length
                    }
                    continue
                }
                if (innerElement.children.length > 0 && getTargetOffset(innerElement)) {
                    return true
                }
            }
        }
        const lineSpan = pre.firstChild as HTMLElement
        if (!lineSpan) {
            console.error('Could not find line element.')
            return
        }
        if (!getTargetOffset(lineSpan)) {
            return
        }
        const type = lineData ? lineData.lineType : 'CONTEXT'
        // If using a split editor check if the current line is a split diff or not.
        const useBase = lineSpan.closest('.side-by-side-diff-editor-from')
        rev = rev || this.props.revState.headRev

        // Check which part of the diff it is.
        if (type !== 'ADDED' || useBase) {
            rev = this.props.revState.baseRev
        }
        const args = {
            filePath: file,
            commitID: rev,
            repoPath,
            position: { character: characterOffset, line: lineNumber },
        } as AbsoluteRepoFilePosition
        /**
         * For every position, emits an Observable with new values for the `hoverOrError` state.
         * This is a higher-order Observable (Observable that emits Observables).
         */
        // Fetch the hover for that position
        const hoverFetch = this.props.simpleProviderFns.fetchHover(args).pipe(
            catchError(error => {
                this.updateTooltipStyle()
                if (error && error.code === EMODENOTFOUND) {
                    return [undefined]
                }
                return [asError(error)]
            })
        )
        // 1. Show a loader if the hover fetch hasn't returned after 100ms
        // 2. Show the hover once it returned
        this.subscriptions.add(
            merge(
                of(LOADING).pipe(
                    delay(LOADER_DELAY),
                    takeUntil(hoverFetch)
                ),
                hoverFetch
            ).subscribe(hoverOrError => {
                if (!hoverOrError || (hoverOrError as any).length === 0) {
                    return
                }
                const data = {
                    target: this.props.element,
                    ctx: args,
                    contents: (hoverOrError as any).contents,
                    range: (hoverOrError as any).range,
                    asyncDefUrl: true,
                    loading: hoverOrError === LOADING,
                } as TooltipData
                this.setState(
                    () => ({ tooltipData: data }),
                    () => {
                        updateTooltip(data, this.props.docked || false, this.tooltipActions(args), false)
                        this.updateTooltipStyle()
                    }
                )
            })
        )
    }

    private getLineNumberFromGutter(pre: HTMLPreElement): number | undefined {
        const lineGutter = pre.previousSibling as HTMLElement | undefined
        if (!lineGutter || lineGutter.nodeType === Node.TEXT_NODE) {
            console.warn('Err: Line Gutter not found.', pre)
            return undefined
        }
        const lineNumberElement = lineGutter.querySelector('a')
        if (!lineNumberElement) {
            console.warn('Err: No lineNumberElement found.')
            return undefined
        }
        return Number(lineNumberElement.getAttribute('data-line-number'))
    }

    private getFileCommitStateFromToolbar(): { file: string; rev?: string } | undefined {
        const detailContainer = (this.props.element.closest('.file-content') ||
            this.props.element.closest('.detail')) as HTMLDivElement
        if (!detailContainer) {
            return undefined
        }
        const fileToolbar = detailContainer.querySelector('.file-toolbar')
        if (!fileToolbar) {
            return undefined
        }
        const breadcrumbs = fileToolbar.querySelector('.breadcrumbs') as HTMLDivElement
        const stub = fileToolbar.querySelector('.stub') as HTMLAnchorElement
        if (breadcrumbs && stub) {
            try {
                const stubUrl = new URL(stub.href)
                const parsedUrl = stubUrl.pathname.split('/')
                if (parsedUrl.length >= 8 && parsedUrl[7] === 'commits') {
                    return { file: breadcrumbs.innerText, rev: parsedUrl[8] }
                }
            } catch {
                /** noop */
            }
            return { file: breadcrumbs.innerText }
        }
    }

    private getLineNumberForPullRequest():
        | { lineNumber: number; lineType: 'CONTEXT' | 'ADDED' | 'REMOVED' }
        | undefined {
        const pre = this.props.element.closest('pre')
        if (!pre || !pre.firstChild) {
            console.warn('Err: No pre element.')
            return undefined
        }
        const lineGutter = pre.previousSibling as HTMLElement | undefined
        if (!lineGutter || lineGutter.nodeType === Node.TEXT_NODE) {
            console.warn('Err: Line Gutter not found.', pre)
            return undefined
        }
        // Get the line number marker.
        const lineNumberMarker = lineGutter.querySelector('.line-number-marker') as HTMLDivElement
        if (!lineNumberMarker) {
            return undefined
        }
        const lineType = lineNumberMarker.getAttribute('data-line-type') as 'CONTEXT' | 'ADDED' | 'REMOVED'
        const lineNumberToElement = lineGutter.querySelector('.line-number-to') as HTMLDivElement
        let lineNumber = lineNumberMarker.getAttribute('data-line-number')
        if (!lineType) {
            console.warn('Err: No lineType', lineNumberMarker)
            return undefined
        }
        if (!lineNumber) {
            console.warn('Err: No lineNumber', lineNumberMarker)
            return undefined
        }
        if (lineType === 'ADDED' && lineNumberToElement) {
            lineNumber = lineNumberToElement.innerText
        }
        return { lineNumber: Number(lineNumber), lineType }
    }

    private tooltipActions = (ctx: AbsoluteRepoFilePosition) => ({
        definition: this.handleGoToDefinition,
        references: this.handleFindReferences,
        dismiss: this.handleDismiss,
    })

    private handleFindReferences = (ctx: AbsoluteRepoFilePosition) => (e: MouseEvent) => {
        eventLogger.logCodeIntelligenceEvent()
    }

    private handleGoToDefinition = (defCtx: AbsoluteRepoFilePosition) => (e: MouseEvent) => {
        eventLogger.logCodeIntelligenceEvent()
        e.preventDefault()
        fetchJumpURL(this.props.simpleProviderFns.fetchDefinition, defCtx).subscribe(defUrl => {
            if (!defUrl) {
                return
            }
            if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
                window.open(defUrl, '_blank')
                return
            }
            const { filePath, repoPath, position, rev } = parseBrowserRepoURL(defUrl)
            // Check if it's actually hosted here. If it's not open on sourcegraph / or at least GitHub.
            if (repoPath.startsWith('github.com')) {
                const line = position ? `#L${position.line}${position.character ? ':' + position.character : ''}` : ''
                const url = `https://${repoPath}/blob/${rev || 'HEAD'}/${filePath}${line}`
                window.location.href = url
                return
            }

            // If it's the same repository and file then just scroll to the element.
            const splitPath = repoPath.split('/')
            const compareUrl = `${window.location.origin}/projects/${splitPath[1]}/repos/${
                splitPath[2]
            }/browse/${filePath}?at=${rev}`
            const url = `${compareUrl}${position ? `#${position.line}` : ''}`
            if (compareUrl === `${window.location.origin}${window.location.pathname}${window.location.search}`) {
                if (position) {
                    scrollToLine(position.line)
                }
            }
            window.location.href = url
        })
    }

    private updateTooltipStyle = () => {
        if (!this.tooltipRef) {
            return
        }
        if (!this.state.tooltipData || !this.state.tooltipData.contents) {
            this.tooltipRef.style.display = 'none'
        } else {
            this.tooltipRef.style.display = 'inline-block'
        }
    }

    private handleDismiss = () => {
        hideTooltip()
    }

    private handleAnchorClick = () => {
        if (!this.state.tooltipData) {
            return
        }
        updateTooltip(this.state.tooltipData, true, this.tooltipActions(this.state.tooltipData.ctx), false)
    }

    public render(): JSX.Element {
        const { offsetWidth, offsetHeight, offsetLeft, offsetTop } = this.props.element
        return (
            <span
                ref={e => (this.tooltipRef = e)}
                id="tooltip-portal"
                onClick={this.handleAnchorClick}
                style={{
                    cursor: 'pointer',
                    backgroundColor: 'rgba(3, 102, 214, 0.3)',
                    width: offsetWidth,
                    height: offsetHeight,
                    display: 'none',
                    position: 'absolute',
                    left: offsetLeft,
                    top: offsetTop,
                }}
            />
        )
    }
}
