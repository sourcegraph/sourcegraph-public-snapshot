import EmailOpenOutlineIcon from 'mdi-react/EmailOpenOutlineIcon'
import PencilBoxIcon from 'mdi-react/PencilBoxIcon'
import PlusBoxIcon from 'mdi-react/PlusBoxIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import WebhookIcon from 'mdi-react/WebhookIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'

interface Props {
    areaURL: string

    className?: string
}

/**
 * The sidebar for the thread actions area.
 */
export const ThreadActionsAreaSidebar: React.FunctionComponent<Props> = ({ areaURL, className = '' }) => (
    <div className={`thread-actions-area-sidebar d-flex flex-column ${className}`}>
        <div className="card">
            <header className="card-header d-flex align-items-center justify-content-between">
                Actions
                <button type="button" className="btn btn-icon text-decoration-none">
                    <PlusBoxIcon className="icon-inline" />
                </button>
            </header>
            <div className="list-group list-group-flush">
                <NavLink
                    to={`${areaURL}/pull-requests`}
                    className="list-group-item list-group-item-action p-2"
                    activeClassName="active"
                >
                    <SourcePullIcon className="icon-inline" /> Pull requests
                </NavLink>
                <NavLink
                    to={`${areaURL}/commit-statuses`}
                    className="list-group-item list-group-item-action p-2"
                    activeClassName="active"
                >
                    <SourceCommitIcon className="icon-inline" /> Commit statuses
                </NavLink>
                <NavLink
                    to={`${areaURL}/slack`}
                    className="list-group-item list-group-item-action p-2"
                    activeClassName="active"
                >
                    <SlackIcon className="icon-inline" /> Slack
                </NavLink>
                <NavLink
                    to={`${areaURL}/email`}
                    className="list-group-item list-group-item-action p-2"
                    activeClassName="active"
                >
                    <EmailOpenOutlineIcon className="icon-inline" /> Email
                </NavLink>
                <NavLink
                    to={`${areaURL}/editor`}
                    className="list-group-item list-group-item-action p-2"
                    activeClassName="active"
                >
                    <PencilBoxIcon className="icon-inline" /> Editor
                </NavLink>
                <NavLink
                    to={`${areaURL}/webhooks`}
                    className="list-group-item list-group-item-action p-2"
                    activeClassName="active"
                >
                    <WebhookIcon className="icon-inline" /> Webhooks
                </NavLink>
            </div>
        </div>
    </div>
)
