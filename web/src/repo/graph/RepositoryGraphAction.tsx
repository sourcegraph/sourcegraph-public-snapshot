import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { encodeRepoRev } from '../../util/url'

/**
 * A repository header action links to the repository graph area.
 */
export class RepositoryGraphAction extends React.PureComponent<{ repo: string; rev: string | undefined }> {
    public render(): JSX.Element | null {
        return (
            <NavLink
                to={`/${encodeRepoRev(this.props.repo, this.props.rev)}/-/graph`}
                className="composite-container__header-action"
                activeClassName="composite-container__header-action-active"
                data-tooltip="Repository graph"
            >
                <GlobeIcon className="icon-inline" />
                <span className="composite-container__header-action-text">Graph</span>
            </NavLink>
        )
    }
}
