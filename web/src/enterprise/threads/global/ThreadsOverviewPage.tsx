import H from 'history'
import { upperFirst } from 'lodash'
import React from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { WithQueryParameter } from '../components/withQueryParameter/WithQueryParameter'
import { ThreadsList } from '../list/ThreadsList'
import { ThreadsListHeader } from '../list/ThreadsListHeader'
import { ThreadsListHeaderFilterButtonDropdown } from '../list/ThreadsListHeaderFilterButtonDropdown'
import { threadsQueryWithValues } from '../url'
import { ThreadsAreaContext } from './ThreadsArea'

interface Props extends ThreadsAreaContext {
    history: H.History
    location: H.Location
}

/**
 * The threads overview page.
 */
export const ThreadsOverviewPage: React.FunctionComponent<Props> = props => (
    <div className="threads-overview-page container mt-3">
        <PageTitle title={upperFirst(props.type)} />
        <WithQueryParameter
            defaultQuery={threadsQueryWithValues('', { is: [props.type.toLowerCase(), 'open'] })}
            history={props.history}
            location={props.location}
        >
            {({ query, onQueryChange }) => (
                <>
                    <ThreadsListHeader {...props} query={query} onQueryChange={onQueryChange} />
                    <ThreadsList
                        {...props}
                        query={query}
                        onQueryChange={onQueryChange}
                        itemCheckboxes={true}
                        rightHeaderFragment={
                            <>
                                {' '}
                                <ThreadsListHeaderFilterButtonDropdown
                                    header="Filter by who's assigned"
                                    items={['sqs (you)', 'ekonev', 'jleiner', 'ziyang', 'kting7', 'ffranksena']}
                                >
                                    Assignee
                                </ThreadsListHeaderFilterButtonDropdown>
                                <ThreadsListHeaderFilterButtonDropdown
                                    header="Filter by label"
                                    items={[
                                        'perf',
                                        'tech-lead',
                                        'services',
                                        'bugs',
                                        'build',
                                        'noisy',
                                        'security',
                                        'appsec',
                                        'infosec',
                                        'compliance',
                                        'docs',
                                    ]}
                                >
                                    Labels
                                </ThreadsListHeaderFilterButtonDropdown>
                                <ThreadsListHeaderFilterButtonDropdown
                                    header="Sort by"
                                    items={['Priority', 'Most recently updated', 'Least recently updated']}
                                >
                                    Sort
                                </ThreadsListHeaderFilterButtonDropdown>
                            </>
                        }
                    />
                </>
            )}
        </WithQueryParameter>
    </div>
)
