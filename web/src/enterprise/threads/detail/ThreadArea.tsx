import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { RepositoryThreadsAreaContext } from '../repository/RepositoryThreadsArea'
import { ThreadRepositoriesList } from './repositories/ThreadRepositoriesList'
import { ThreadOverview } from './ThreadOverview'
import { ThreadThreadsListPage } from './threads/ThreadThreadsListPage'
import { useThreadByID } from './useThreadByID'

export interface ThreadAreaContext
    extends Pick<RepositoryThreadsAreaContext, Exclude<keyof RepositoryThreadsAreaContext, 'namespace'>> {
    /** The thread ID. */
    threadID: GQL.ID

    /** The thread, queried from the GraphQL API. */
    thread: GQL.IThread

    /** Called to refresh the thread. */
    onThreadUpdate: () => void

    location: H.Location
    history: H.History
}

interface Props
    extends Pick<ThreadAreaContext, Exclude<keyof ThreadAreaContext, 'thread' | 'onThreadUpdate'>>,
        RouteComponentProps<never> {
    header: React.ReactFragment
}

const LOADING = 'loading' as const

const PAGE_CLASS_NAME = 'container mt-4'

/**
 * The area for a single thread.
 */
export const ThreadArea: React.FunctionComponent<Props> = ({
    header,
    threadID,
    setBreadcrumbItem,
    match,
    ...props
}) => {
    const [threadOrError, onThreadUpdate] = useThreadByID(threadID)

    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem(
                threadOrError !== LOADING && threadOrError !== null && !isErrorLike(threadOrError)
                    ? { text: threadOrError.name, to: threadOrError.url }
                    : undefined
            )
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [threadOrError, setBreadcrumbItem])

    if (threadOrError === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (threadOrError === null) {
        return <HeroPage icon={AlertCircleIcon} title="Thread not found" />
    }
    if (isErrorLike(threadOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={threadOrError.message} />
    }

    const context: ThreadAreaContext = {
        ...props,
        threadID,
        thread: threadOrError,
        onThreadUpdate,
        setBreadcrumbItem,
    }

    return (
        <>
            <style>{`.user-area-header, .org-header { display: none; } .org-area > .container, .user-area > .container { margin: unset; margin-top: unset !important; width: unset; padding: unset; } /* TODO!(sqs): hack */`}</style>
            <OverviewPagesArea<ThreadAreaContext>
                context={context}
                header={header}
                overviewComponent={ThreadOverview}
                pages={[
                    {
                        title: 'Threads',
                        path: '',
                        render: () => <ThreadThreadsListPage {...context} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'Commits',
                        path: '/commits',
                        render: () => <ThreadRepositoriesList {...context} className={PAGE_CLASS_NAME} />,
                    },
                    // { title: 'Changes', path: '/changes', render: () => <ThreadChangesListPage {...context} /> },
                ]}
                location={props.location}
                match={match}
            />
        </>
    )
}
