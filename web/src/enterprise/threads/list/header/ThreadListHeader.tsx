import H from 'history'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import TagOutlineIcon from 'mdi-react/TagOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { QueryParameterProps } from '../../../../util/useQueryParameter'
import { ThreadsAreaContext } from '../global/ThreadsArea'
import { threadNoun } from '../util'
import { ThreadsListFilter } from './ThreadsListFilter'
import { ThreadListButtonDropdownFilter } from './ThreadListFilterButtonDropdown'

interface Props extends QueryParameterProps, Pick<ThreadsAreaContext, 'type'> {
    location: H.Location
}

/**
 * The header for the list of threads.
 * TODO!(sqs) use this
 */
export const ThreadListHeader: React.FunctionComponent<Props> = ({
    type,
    query,
    onQueryChange,
    locationWithQuery,
    location,
}) => (
    <div className="d-flex justify-content-between align-items-start">
        <div className="flex-1 mr-5 d-flex">
            <div className="flex-1 mb-3 mr-2">
                <ThreadsListFilter
                    value={query}
                    onChange={onQueryChange}
                    beforeInputFragment={
                        <div className="input-group-prepend">
                            <ThreadListButtonDropdownFilter locationWithQuery={locationWithQuery} />
                        </div>
                    }
                />
            </div>
            <div className="btn-group mb-3 mr-4" role="group">
                <Link to={`${location.pathname}/-/manage/activity`} className="btn btn-link">
                    <HistoryIcon className="icon-inline mr-1" /> Activity
                </Link>
                <Link to={`${location.pathname}/labels`} className="btn btn-link">
                    <TagOutlineIcon className="icon-inline mr-1" /> Labels{' '}
                    <span className="badge badge-secondary ml-1">9</span>
                </Link>
            </div>
        </div>
        <div className="btn-group" role="group">
            <Link to={`${location.pathname}/-/manage`} className="btn btn-link">
                <SettingsIcon className="icon-inline mr-1" /> Manage{' '}
            </Link>
            <Link to={`${location.pathname}/-/new`} className="btn btn-success">
                New {threadNoun(type)}
            </Link>
        </div>
    </div>
)
