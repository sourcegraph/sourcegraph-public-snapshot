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
import { ChangesetCommitsList } from './commits/ChangesetCommitsList'
import { ThreadOverview } from './ThreadOverview'
import { useThreadByIDInRepository } from './useThreadByID'

export interface ThreadAreaContext
    extends Pick<RepositoryThreadsAreaContext, Exclude<keyof RepositoryThreadsAreaContext, 'repository'>> {
    /** The thread, queried from the GraphQL API. */
    thread: GQL.IThread

    /** Called to refresh the thread. */
    onThreadUpdate: () => void

    location: H.Location
    history: H.History
}

interface Props
    extends Pick<ThreadAreaContext, Exclude<keyof ThreadAreaContext, 'thread' | 'onThreadUpdate'>>,
        RouteComponentProps<{}> {
    /**
     * The thread ID in its repository (i.e., the `Thread.idInRepository` GraphQL API field).
     */
    threadIDInRepository: GQL.IThread['idInRepository']

    header: React.ReactFragment
}

const LOADING = 'loading' as const

const PAGE_CLASS_NAME = 'container mt-4'

/**
 * The area for a single thread.
 */
export const ThreadArea: React.FunctionComponent<Props> = ({
    header,
    threadIDInRepository,
    setBreadcrumbItem,
    match,
    ...props
}) => {
    const [thread, onThreadUpdate] = useThreadByIDInRepository(props.repo.id, threadIDInRepository)

    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem(
                thread !== LOADING && thread !== null && !isErrorLike(thread)
                    ? { text: `#${thread.idInRepository}`, to: thread.url }
                    : undefined
            )
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [thread, setBreadcrumbItem])

    if (thread === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (thread === null) {
        return <HeroPage icon={AlertCircleIcon} title="Thread not found" />
    }
    if (isErrorLike(thread)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={thread.message} />
    }

    const context: ThreadAreaContext = {
        ...props,
        thread,
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
                        title: 'Commits',
                        path: '/commits',
                        render: () => <ChangesetCommitsList {...context} className={PAGE_CLASS_NAME} />,
                    },
                    // { title: 'Changes', path: '/changes', render: () => <ThreadChangesListPage {...context} /> },
                ]}
                location={props.location}
                match={match}
            />
        </>
    )
}
