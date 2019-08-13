import H from 'history'
import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThreadListFilterContext } from '../../../threads/list/header/ThreadListFilterDropdownButton'
import { ThreadListHeaderCommonFilters } from '../../../threads/list/header/ThreadListHeader'
import { ThreadList } from '../../../threads/list/ThreadList'
import { ThreadsListFilter } from '../../../threadsOLD/list/ThreadsListFilter'
import { ThreadsListButtonDropdownFilter } from '../../../threadsOLD/list/ThreadsListFilterButtonDropdown'
import { ThreadsListHeaderFilterButtonDropdown } from '../../../threadsOLD/list/ThreadsListHeaderFilterButtonDropdown'

const LOADING = 'loading' as const

interface Props extends QueryParameterProps {
    threads: typeof LOADING | GQL.IThreadConnection | ErrorLike
    campaign: Pick<GQL.ICampaign, 'id'>
    action: React.ReactFragment

    className?: string
    location: H.Location
    history: H.History
}

export const CampaignThreadList: React.FunctionComponent<Props> = ({
    threads,
    campaign,
    action,
    className = '',
    query,
    onQueryChange,
    ...props
}) => {
    const filterProps: ThreadListFilterContext = {
        threadConnection: threads,
        query,
        onQueryChange,
    }
    return (
        <div className={`campaign-thread-list ${className}`}>
            <header className="d-flex justify-content-between align-items-start">
                <div className="flex-1 mr-2 d-flex">
                    <div className="flex-1 mb-3 mr-2">
                        <ThreadsListFilter
                            value={query}
                            onChange={onQueryChange}
                            beforeInputFragment={
                                <div className="input-group-prepend">
                                    <ThreadsListButtonDropdownFilter />
                                </div>
                            }
                        />
                    </div>
                </div>
                {action}
            </header>
            <ThreadList
                {...props}
                threads={threads}
                query={query}
                onQueryChange={onQueryChange}
                itemCheckboxes={true}
                showRepository={true}
                headerItems={{
                    right: (
                        <>
                            <ThreadListHeaderCommonFilters {...filterProps} />
                        </>
                    ),
                }}
            />
        </div>
    )
}
