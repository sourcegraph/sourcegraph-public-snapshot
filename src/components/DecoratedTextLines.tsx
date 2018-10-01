import * as React from 'react'
import VisibilitySensor from 'react-visibility-sensor'
import * as GQL from '../backend/graphqlschema'
import { highlightNode } from '../util/dom'
import { LinkOrSpan } from './LinkOrSpan'

interface Props {
    /**
     * A CSS class name to add to this component's element.
     */
    className?: string

    /**
     * The code string (or array of lines) to display.
     */
    value: string | string[]

    /**
     * The highlights for the lines.
     */
    highlights?: GQL.IHighlight[]

    /**
     * A list of classes to apply to 1-indexed line numbers.
     */
    lineClasses?: { line: number; className: string; url?: string }[]

    /**
     * Called when the mousedown event is triggered on the element.
     */
    onMouseDown?: () => void
}

interface DecoratedLine {
    value: string
    highlights?: GQL.IHighlight[]
    classNames?: string[]
    url?: string
}

interface State {
    visible: boolean
    lines: DecoratedLine[]
}

/**
 * A simple component for displaying lines of text, with optional
 * highlighted ranges (not syntax highlighting, only e.g. query match
 * highlighting).
 */
export class DecoratedTextLines extends React.PureComponent<Props, State> {
    private tableContainerElement: HTMLElement | null = null

    constructor(props: Props) {
        super(props)
        this.state = {
            ...this.getStateForProps(props),
            visible: false,
        }
    }

    public componentDidMount(): void {
        this.updateHighlights()
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (
            this.props.value !== nextProps.value ||
            this.props.highlights !== nextProps.highlights ||
            this.props.lineClasses !== nextProps.lineClasses
        ) {
            this.setState(this.getStateForProps(nextProps))
        }
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        this.updateHighlights()
    }

    private updateHighlights(): void {
        if (this.state.visible && this.tableContainerElement) {
            const rows = this.tableContainerElement.querySelectorAll('table tr')
            for (const [i, row] of rows.entries()) {
                const elem = row.firstChild as HTMLTableDataCellElement
                const data = this.state.lines[i]
                if (data.highlights && data.highlights.length) {
                    // TODO(sqs): only supports 1 highlight per line
                    const highlight = data.highlights[0]
                    highlightNode(elem, highlight.character, highlight.length)
                }
            }
        }
    }

    private getStateForProps(props: Props): { lines: DecoratedLine[] } {
        const lineValues = typeof props.value === 'string' ? props.value.split('\n') : props.value
        const lines: DecoratedLine[] = lineValues.map(line => ({ value: line }))
        if (props.highlights) {
            for (const highlight of props.highlights) {
                const line = lines[highlight.line - 1]
                if (!line.highlights) {
                    line.highlights = []
                }
                line.highlights.push(highlight)
            }
        }
        if (props.lineClasses) {
            for (const { line: lineNumber, className, url } of props.lineClasses) {
                const line = lines[lineNumber - 1]
                if (!line.classNames) {
                    line.classNames = []
                }
                line.classNames.push(className)
                if (url) {
                    line.url = url
                }
            }
        }
        return { lines }
    }

    public render(): JSX.Element | null {
        return (
            <VisibilitySensor onChange={this.onChangeVisibility} partialVisibility={true}>
                <code className={`decorated-text-lines code-excerpt ${this.props.className || ''}`}>
                    <table ref={this.setTableContainerElement}>
                        <tbody>
                            {this.state.lines.map((line, i) => (
                                <tr key={i} className={line.classNames ? line.classNames.join(' ') : undefined}>
                                    <td className="code" onMouseDown={this.props.onMouseDown}>
                                        <LinkOrSpan to={line.url}>{line.value}</LinkOrSpan>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </code>
            </VisibilitySensor>
        )
    }

    public onChangeVisibility = (isVisible: boolean): void => {
        this.setState({ visible: true })
    }

    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
    }
}
