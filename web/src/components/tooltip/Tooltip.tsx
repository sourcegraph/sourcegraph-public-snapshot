import * as React from 'react'

interface Props {}

interface State {
    subject?: HTMLElement
    top?: number
    left?: number
    content?: string
}

/**
 * A global tooltip displayed for elements containing a `data-tooltip` attribute.
 */
export class Tooltip extends React.PureComponent<Props, State> {
    private static SUBJECT_ATTRIBUTE = 'data-tooltip'
    private static DELAY = 50

    /**
     * Singleton instance, so that other components can call Tooltip.forceUpdate().
     */
    private static INSTANCE: Tooltip | undefined

    public state: State = {}

    private containerRef: HTMLElement | null = null
    private tooltipRef: HTMLElement | null = null
    private _timeout?: number

    /**
     * Forces an update. Other components must call this if they modify their
     * tooltip content while the tooltip is still visible.
     */
    public static forceUpdate(): void {
        if (Tooltip.INSTANCE) {
            Tooltip.INSTANCE.forceUpdate()
        }
    }

    public componentDidMount(): void {
        Tooltip.INSTANCE = this

        document.addEventListener('focusin', this.toggleHint)
        document.addEventListener('mouseover', this.toggleHint)
        document.addEventListener('touchend', this.toggleHint)
        document.addEventListener('click', this.toggleHint)
    }

    public componentDidUpdate(): void {
        if (this.state.subject) {
            const data = this.getStateForSubject(this.state.subject)
            if (data) {
                this.setState(data)
            }
        }
    }

    public componentWillUnmount(): void {
        Tooltip.INSTANCE = undefined

        document.removeEventListener('focusin', this.toggleHint)
        document.removeEventListener('mouseover', this.toggleHint)
        document.removeEventListener('touchend', this.toggleHint)
        document.removeEventListener('click', this.toggleHint)
        if (this._timeout !== undefined) {
            clearTimeout(this._timeout)
        }
    }

    public render(): React.ReactFragment | null {
        const rebasedLeft = (this.state.left || 0) + window.innerWidth
        return (
            <div ref={this.setContainerRef} className="tooltip2__container">
                {this.state.subject && (
                    <div
                        className="tooltip2 tooltip2--bottom"
                        ref={this.setTooltipRef}
                        // tslint:disable-next-line:jsx-ban-props
                        style={{
                            top: this.state.top,
                            left:
                                Math.max(
                                    5 /* min margin */,
                                    Math.min(window.innerWidth - 108 /* 18rem * 6 */ - 5 /* min margin */, rebasedLeft)
                                ) - window.innerWidth,
                        }}
                    >
                        <div className="tooltip2__content">{this.state.content}</div>
                    </div>
                )}
            </div>
        )
    }

    private setContainerRef = (e: HTMLElement | null) => (this.containerRef = e)
    private setTooltipRef = (e: HTMLElement | null) => (this.tooltipRef = e)

    private toggleHint = (e: Event): void => {
        if (this._timeout !== undefined) {
            clearTimeout(this._timeout)
        }

        // As a special case, don't show the tooltip for click events on submit buttons that are probably triggered
        // by the user pressing the enter button. It is not desirable for the tooltip to be shown in that case.
        if (
            e.type === 'click' &&
            (e.target as HTMLElement).tagName === 'BUTTON' &&
            (e.target as HTMLButtonElement).type === 'submit' &&
            (e as MouseEvent).pageX === 0 &&
            (e as MouseEvent).pageY === 0
        ) {
            return
        }

        this._timeout = window.setTimeout(
            () =>
                this.setState(() => ({
                    subject: this.getSubject(e.target as HTMLElement),
                })),
            Tooltip.DELAY
        )
    }

    /**
     * Find the nearest ancestor element to e that contains a tooltip.
     */
    private getSubject = (e: HTMLElement | null): HTMLElement | undefined => {
        while (e) {
            if (e === document.body) {
                break
            }
            if (e.hasAttribute(Tooltip.SUBJECT_ATTRIBUTE)) {
                // If e is not actually attached to the DOM, then abort.
                if (!document.body.contains(e)) {
                    return undefined
                }
                return e
            }
            e = e.parentElement
        }
        return undefined
    }

    public getStateForSubject = (subject: HTMLElement): { content: string; top: number; left: number } | undefined => {
        if (!this.containerRef || !this.tooltipRef || !document.body.contains(subject)) {
            return undefined
        }

        const content = subject.getAttribute(Tooltip.SUBJECT_ATTRIBUTE) || ''

        const {
            top: containerTop,
            left: containerLeft,
            right: containerRight,
        } = this.containerRef.getBoundingClientRect()

        const { width: tooltipWidth } = this.tooltipRef.getBoundingClientRect()

        const {
            top: subjectTop,
            left: subjectLeft,
            width: subjectWidth,
            height: subjectHeight,
        } = subject.getBoundingClientRect()

        const top = subjectHeight
        let left = (subjectWidth - tooltipWidth) / 2

        const outOfBoundsWidth = Math.floor(containerRight - subjectLeft - left - tooltipWidth)
        if (outOfBoundsWidth < 0) {
            left += outOfBoundsWidth
        }

        return {
            content,
            top: top + subjectTop - containerTop,
            left: left + subjectLeft - containerLeft,
        }
    }
}
