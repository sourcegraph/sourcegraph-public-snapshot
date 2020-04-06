import * as H from 'history'
import EyeIcon from 'mdi-react/EyeIcon'
import * as React from 'react'
import { LinkOrButton } from '../../../../../shared/src/components/LinkOrButton'
import { RenderMode } from '../../../../../shared/src/util/url'
import { Tooltip } from '../../../components/tooltip/Tooltip'

interface Props {
    location: H.Location
    mode: RenderMode
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
    public static getModeFromURL(location: H.Location): RenderMode {
        const q = new URLSearchParams(location.search)

        if (!q.has(ToggleRenderedFileMode.URL_QUERY_PARAM)) {
            const isDiscussions = new URLSearchParams(location.hash).get('tab') === 'discussions'
            if (isDiscussions) {
                return 'code'
            }
            return undefined
        }
        return q.get(ToggleRenderedFileMode.URL_QUERY_PARAM) === 'code' ? 'code' : 'rendered' // default to rendered
    }

    /**
     * Returns the URL that displays the blob using the specified mode.
     */
    private getURLForMode(location: H.Location, mode: RenderMode): H.Location {
        const q = new URLSearchParams(location.search)
        if (mode === 'code') {
            q.set(ToggleRenderedFileMode.URL_QUERY_PARAM, mode)
        } else {
            q.delete(ToggleRenderedFileMode.URL_QUERY_PARAM)
        }
        return { ...location, search: q.toString() }
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.mode !== this.props.mode) {
            Tooltip.forceUpdate()
        }
    }

    public render(): JSX.Element | null {
        const otherMode: RenderMode = this.props.mode === 'code' ? 'rendered' : 'code'

        return (
            <LinkOrButton
                to={this.getURLForMode(this.props.location, otherMode)}
                data-tooltip={otherMode === 'code' ? 'Show raw code file' : 'Show formatted file'}
            >
                <EyeIcon className="icon-inline" />{' '}
                <span className="d-none d-lg-inline">{otherMode === 'code' ? 'Raw' : 'Formatted'}</span>
            </LinkOrButton>
        )
    }
}
