import * as React from 'react'
import { Link } from 'react-router-dom'
import { highlightNode } from '../util/dom'
import { toPositionOrRangeHash } from '../util/url'

interface Props {
    urlWithoutPosition: string
    items: {
        line: number
        preview: string
        highlightRanges: {
            start: number
            highlightLength: number
        }[]
    }[]
    onSelect: () => void
}

export class CodeExcerpt2 extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const maxDigits = this.props.items.reduce((n, { line }) => {
            const digits = (line + 1).toString().length
            return digits > n ? digits : n
        }, 1)
        return (
            <pre className="file-match__code">
                {this.props.items.map(({ line, preview, highlightRanges }, i) => (
                    <code
                        data-line={(line + 1).toString().padStart(maxDigits)}
                        key={i}
                        ref={e => e && highlightNode(e, highlightRanges[0].start, highlightRanges[0].highlightLength)}
                    >
                        <Link
                            to={`${this.props.urlWithoutPosition}${toPositionOrRangeHash({
                                position: { line: line + 1, character: highlightRanges[0].start + 1 },
                            })}`}
                            key={i}
                            onClick={this.props.onSelect}
                        >
                            {preview}
                        </Link>
                    </code>
                ))}
            </pre>
        )
    }
}
