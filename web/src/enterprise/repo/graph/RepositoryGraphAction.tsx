import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import * as React from 'react'
import { ActionItem } from '../../../components/ActionItem'
import { encodeRepoRev } from '../../../util/url'

/**
 * A repository header action links to the repository graph area.
 */
export class RepositoryGraphAction extends React.PureComponent<{
    repo: string
    rev: string | undefined
}> {
    public render(): JSX.Element | null {
        return (
            <ActionItem
                to={`/${encodeRepoRev(this.props.repo, this.props.rev)}/-/graph`}
                data-tooltip="Repository graph"
            >
                <GlobeIcon className="icon-inline" /> <span className="d-none d-lg-inline">Graph</span>
            </ActionItem>
        )
    }
}
