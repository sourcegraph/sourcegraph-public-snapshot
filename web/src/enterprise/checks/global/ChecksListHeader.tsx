import H from 'history'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import TagOutlineIcon from 'mdi-react/TagOutlineIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ChecksListFilter } from './ChecksListFilter'

interface Props {
    location: H.Location
}

/**
 * The header for the list of checks.
 */
export class ChecksListHeader extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="d-flex justify-content-between align-items-start">
                <div className="flex-1 mr-5 d-flex">
                    <div className="flex-1 mb-3 mr-2">
                        <ChecksListFilter className="" value="is:open sort:priority" />
                    </div>
                    <div className="btn-group mb-3 mr-4" role="group">
                        <Link to={`${this.props.location.pathname}/-/manage/activity`} className="btn btn-outline-link">
                            <HistoryIcon className="icon-inline" /> Activity
                        </Link>
                        <Link to={`${this.props.location.pathname}/labels`} className="btn btn-outline-link">
                            <TagOutlineIcon className="icon-inline" /> Labels{' '}
                            <span className="badge badge-secondary">9</span>
                        </Link>
                    </div>
                </div>
                <div className="btn-group" role="group">
                    <Link to={`${this.props.location.pathname}/-/manage`} className="btn btn-outline-link">
                        <SettingsIcon className="icon-inline" /> Manage{' '}
                    </Link>
                    <Link to={`${this.props.location.pathname}/-/manage/new`} className="btn btn-success">
                        New check
                    </Link>
                </div>
            </div>
        )
    }
}
