import formatDistance from 'date-fns/formatDistance'
import * as React from 'react'

interface Props {
    /** The date string (RFC 3339). */
    date: string
}

/**
 * Displays a date's relative time ("... ago") and shows the full date on hover. Re-renders
 * periodically to ensure the relative time string is up-to-date.
 */
export class Timestamp extends React.PureComponent<Props> {
    private static RERENDER_INTERVAL_MSEC = 7000

    private intervalHandle: number

    public componentDidMount(): void {
        this.intervalHandle = window.setInterval(() => this.forceUpdate(), Timestamp.RERENDER_INTERVAL_MSEC)
    }

    public componentWillUnmount(): void {
        window.clearInterval(this.intervalHandle)
    }

    public render(): JSX.Element {
        return (
            <span className="timestamp" data-tooltip={this.props.date}>
                {formatDistance(this.props.date, new Date(), { addSuffix: true, includeSeconds: true })}
            </span>
        )
    }
}
