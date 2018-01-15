import ViewIcon from '@sourcegraph/icons/lib/View'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Tooltip } from '../../components/tooltip/Tooltip'
import { eventLogger } from '../../tracking/eventLogger'

/**
 * The file rendering modes.
 */
export type FileRenderMode = 'code' | 'rendered'

interface Props {
    location: H.Location
    mode: FileRenderMode
}

/**
 * A repository header action that toggles between showing a rendered file and the file's original
 * source, for files that can be rendered (such as Markdown files).
 */
export class ToggleRenderedFileMode extends React.PureComponent<Props> {
    private static URL_QUERY_PARAM = 'view'

    /**
     * Reports whether the location's URL displays the blob as rendered or source.
     */
    public static getModeFromURL(location: H.Location): FileRenderMode {
        const q = new URLSearchParams(location.search)
        return q.get(ToggleRenderedFileMode.URL_QUERY_PARAM) === 'code' ? 'code' : 'rendered' // default to rendered
    }

    /**
     * Returns the URL that displays the blob using the specified mode.
     */
    private static getURLForMode(location: H.Location, mode: FileRenderMode): { search: string } {
        const q = new URLSearchParams(location.search)
        if (mode === 'code') {
            q.set(ToggleRenderedFileMode.URL_QUERY_PARAM, mode)
        } else {
            q.delete(ToggleRenderedFileMode.URL_QUERY_PARAM)
        }
        return { search: q.toString() }
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.mode !== this.props.mode) {
            Tooltip.forceUpdate()
        }
    }

    public render(): JSX.Element | null {
        const otherMode: FileRenderMode = this.props.mode === 'code' ? 'rendered' : 'code'

        return (
            <Link
                to={ToggleRenderedFileMode.getURLForMode(this.props.location, otherMode)}
                className="composite-container__header-action"
                onClick={this.onClick}
                data-tooltip={otherMode === 'code' ? 'Show raw code file' : 'Show formatted file'}
            >
                <ViewIcon className="icon-inline" />
                <span className="composite-container__header-action-text">
                    {otherMode === 'code' ? 'Raw' : 'Formatted'}
                </span>
            </Link>
        )
    }

    private onClick: React.MouseEventHandler<HTMLAnchorElement> = () => {
        eventLogger.log('ViewButtonClicked')
    }
}
