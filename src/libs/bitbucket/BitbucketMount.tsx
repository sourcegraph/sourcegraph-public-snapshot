import * as React from 'react'
import { createPortal } from 'react-dom'
import { fromEvent, merge, Subscription } from 'rxjs'
import { debounceTime } from 'rxjs/operators'
import { SimpleProviderFns } from '../../shared/backend/lsp'
import { resolveParentRev } from '../../shared/repo/backend'
import { convertNode, createTooltips, hideTooltip, isTooltipDocked } from '../../shared/repo/tooltips'
import { TooltipPortal } from './TooltipPortal'
import { BitbucketState, getRevisionState } from './utils/util'

interface Props {
    container: HTMLElement
    bitbucketState: BitbucketState
    simpleProviderFns: SimpleProviderFns
}

interface State {
    element?: HTMLElement
    event?: MouseEvent
    docked?: boolean
    revState?: {
        baseRev: string
        headRev: string
    }
}

export class BitbucketMount extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            element: undefined,
            event: undefined,
            docked: false,
            revState: undefined,
        }
    }

    public componentWillUnmount(): void {
        hideTooltip()
        this.subscriptions.unsubscribe()
    }

    public componentDidMount(): void {
        const { repository } = this.props.bitbucketState
        if (!repository) {
            return
        }
        const repoPath = `${window.location.hostname}/${repository.project.key}/${repository.slug}`
        // get the current revision details.
        const revState = getRevisionState(this.props.bitbucketState)
        if (!revState) {
            return
        }
        const closestSection = this.props.container.closest('section')
        if (!closestSection) {
            console.error('could not find closest section for event listeners.')
            return
        }
        createTooltips()
        const { baseRev, headRev } = revState
        if (baseRev !== undefined) {
            this.setState(() => ({ revState: { headRev, baseRev } }))
        } else {
            this.subscriptions.add(
                resolveParentRev({ repoPath, rev: revState.headRev! }).subscribe(rev => {
                    this.setState(() => ({ revState: { headRev, baseRev: rev } }))
                })
            )
        }
        this.subscriptions.add(
            merge(fromEvent(closestSection, 'mouseover'), fromEvent(closestSection, 'click'))
                .pipe(debounceTime(50))
                .subscribe((e: MouseEvent) => {
                    this.handleTokens(e)
                })
        )
    }

    private handleTokens = (e: MouseEvent) => {
        if (e.type === 'mouseover') {
            if (isTooltipDocked()) {
                return
            }
        }
        let element = document.elementFromPoint(e.x, e.y) as HTMLElement
        if (!element) {
            return
        }
        // Check if it's already a wrapped node. If it is already a warpped node then go on.
        if (!element.classList.contains('wrapped-node')) {
            const isCodeLine = element.closest('.CodeMirror-line') as HTMLElement
            if (!isCodeLine) {
                if (this.state.element) {
                    this.setState(() => ({ element: undefined, event: undefined, docked: false }))
                }
                return
            }
            // Double clicking on a line causes the line to re-render without the annotations.
            if (isCodeLine && !isCodeLine.querySelector('.annotated')) {
                const hiddenMarker = document.createElement('span')
                hiddenMarker.classList.add('annotated')
                hiddenMarker.style.visibility = 'hidden'
                hiddenMarker.style.display = 'none'
                isCodeLine.appendChild(hiddenMarker)
                convertNode(isCodeLine)
            }
            element = document.elementFromPoint(e.x, e.y) as HTMLElement
            if (!element) {
                return
            }
        }
        if (element.classList.contains('wrapped-node')) {
            if (!element.textContent || !element.textContent.trim().length) {
                if (this.state.element) {
                    this.setState(() => ({ element: undefined, event: undefined, docked: false }))
                }
                return
            }
            // Get line and character offset.
            this.setState({ element, event: e, docked: e.type === 'click' })
            // Single click we set selection.
        } else {
            if (element === this.state.element || element.id === 'tooltip-portal') {
                return
            }
            this.setState(() => ({ element: undefined, event: undefined, docked: false }))
        }
    }

    public render(): JSX.Element | null {
        if (!this.state.element || !this.state.revState) {
            return null
        }

        return (
            <span>
                {createPortal(
                    <TooltipPortal
                        revState={this.state.revState}
                        bitbucketState={this.props.bitbucketState}
                        event={this.state.event}
                        element={this.state.element}
                        docked={this.state.docked}
                        simpleProviderFns={this.props.simpleProviderFns}
                    />,
                    this.state.element
                )}
            </span>
        )
    }
}
