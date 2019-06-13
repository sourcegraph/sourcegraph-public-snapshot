import H from 'history'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import TagOutlineIcon from 'mdi-react/TagOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { QueryParameterProps } from '../../threads/components/withQueryParameter/WithQueryParameter'

interface Props extends QueryParameterProps {
    location: H.Location
    className?: string
}

/**
 * The header for the list of changes.
 */
export const ChangesListHeader: React.FunctionComponent<Props> = ({ location, className = '' }) => (
    <div className={`d-flex justify-content-between align-items-start ${className}`}>
        <div className="flex-1 mr-5 d-flex">
            <div className="btn-group mb-3 mr-4" role="group">
                <Link to={`${location.pathname}/-/manage/activity`} className="btn btn-link">
                    <HistoryIcon className="icon-inline" /> Activity
                </Link>
                <Link to={`${location.pathname}/labels`} className="btn btn-link">
                    <TagOutlineIcon className="icon-inline" /> Labels <span className="badge badge-secondary">9</span>
                </Link>
            </div>
        </div>
        <div className="btn-group" role="group">
            <Link to={`${location.pathname}/-/manage`} className="btn btn-link">
                <SettingsIcon className="icon-inline" /> Manage{' '}
            </Link>
        </div>
    </div>
)
