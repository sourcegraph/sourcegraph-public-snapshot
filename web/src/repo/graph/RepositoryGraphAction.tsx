import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { encodeRepoRev } from '../../util/url'

/**
 * A repository header action links to the repository graph area.
 */
export class RepositoryGraphAction extends React.PureComponent<{ repo: string; rev: string | undefined }> {
    public render(): JSX.Element | null {
        return (
            <Link
                to={`/${encodeRepoRev(this.props.repo, this.props.rev)}/-/graph`}
                className="composite-container__header-action"
                data-tooltip="Repository graph"
            >
                <GlobeIcon className="icon-inline" />
                <span className="composite-container__header-action-text">Graph</span>
            </Link>
        )
    }
}
