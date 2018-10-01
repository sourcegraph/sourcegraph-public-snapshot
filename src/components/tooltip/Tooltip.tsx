import * as React from 'react'
import { Tooltip as BootstrapTooltip } from 'reactstrap'

interface Props {}

interface State {
    subject?: HTMLElement
    subjectSeq: number
    content?: string
}

/**
 * A global tooltip displayed for elements containing a `data-tooltip` attribute.
 */
export class Tooltip extends React.PureComponent<Props, State> {
    private static SUBJECT_ATTRIBUTE = 'data-tooltip'

    /**
     * Singleton instance, so that other components can call Tooltip.forceUpdate().
     */
    private static INSTANCE: Tooltip | undefined

    public state: State = { subjectSeq: 0 }

    /**
     * Forces an update of the tooltip content. Other components must call this if they modify their tooltip
     * content while the tooltip is still visible.
     */
    public static forceUpdate(): void {
        if (Tooltip.INSTANCE) {
            Tooltip.INSTANCE.updateContent()
        }
    }

    public componentDidMount(): void {
        Tooltip.INSTANCE = this

        document.addEventListener('focusin', this.toggleHint)
        document.addEventListener('mouseover', this.toggleHint)
        document.addEventListener('touchend', this.toggleHint)
        document.addEventListener('click', this.toggleHint)
    }

    public componentWillUnmount(): void {
        Tooltip.INSTANCE = undefined

        document.removeEventListener('focusin', this.toggleHint)
        document.removeEventListener('mouseover', this.toggleHint)
        document.removeEventListener('touchend', this.toggleHint)
        document.removeEventListener('click', this.toggleHint)
    }

    public render(): React.ReactFragment | null {
        return this.state.subject ? (
            <BootstrapTooltip
                // Set key prop to work around a bug where quickly mousing between 2 elements with tooltips
                // displays the 2nd element's tooltip as still pointing to the first.
                key={this.state.subjectSeq}
                className="tooltip"
                isOpen={true}
                target={this.state.subject}
                placement="auto"
            >
                {this.state.content}
            </BootstrapTooltip>
        ) : null
    }

    private toggleHint = (event: Event): void => {
        const subject = this.getSubject(event)
        this.setState(prevState => ({
            subject,
            subjectSeq: prevState.subject === subject ? prevState.subjectSeq : prevState.subjectSeq + 1,
            content: subject ? this.getContent(subject) : undefined,
        }))
    }

    private updateContent = () => {
        this.setState(prevState => ({ content: prevState.subject ? this.getContent(prevState.subject) : undefined }))
    }

    /**
     * Find the nearest ancestor element to e that contains a tooltip.
     */
    private getSubject = (event: Event): HTMLElement | undefined => {
        // As a special case, don't show the tooltip for click events on submit buttons that are probably triggered
        // by the user pressing the enter button. It is not desirable for the tooltip to be shown in that case.
        if (
            event.type === 'click' &&
            (event.target as HTMLElement).tagName === 'BUTTON' &&
            (event.target as HTMLButtonElement).type === 'submit' &&
            (event as MouseEvent).pageX === 0 &&
            (event as MouseEvent).pageY === 0
        ) {
            return undefined
        }

        let e: HTMLElement | null = event.target as HTMLElement
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

    private getContent = (subject: HTMLElement): string | undefined => {
        if (!document.body.contains(subject)) {
            return undefined
        }
        return subject.getAttribute(Tooltip.SUBJECT_ATTRIBUTE) || undefined
    }
}
