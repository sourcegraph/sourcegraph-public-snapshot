import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { InfoSidebar, InfoSidebarSection } from '../../../components/infoSidebar/InfoSidebar'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { WithSidebar } from '../../../components/withSidebar/WithSidebar'
import { ThreadDeleteButton } from '../common/ThreadDeleteButton'
import { RepositoryThreadsAreaContext } from '../repository/RepositoryThreadsArea'
import { ThreadActivity } from './activity/ThreadActivity'
import { ThreadCommitsList } from './commits/ThreadCommitsList'
import { ThreadDiagnosticsList } from './diagnostics/ThreadDiagnosticsList'
import { ThreadFileDiffsList } from './fileDiffs/ThreadFileDiffsList'
import { threadSidebarSections } from './sidebar/threadSidebarSections'
import { ThreadOverview } from './ThreadOverview'
import { useThreadByNumberInRepository } from './useThreadByNumberInRepository'

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
     * The thread ID in its repository (i.e., the `Thread.number` GraphQL API field).
     */
    threadNumber: GQL.IThread['number']

    header: React.ReactFragment
}

const LOADING = 'loading' as const

const PAGE_CLASS_NAME = 'container my-5'

/**
 * The area for a single thread.
 */
export const ThreadArea: React.FunctionComponent<Props> = ({
    header,
    threadNumber,
    setBreadcrumbItem,
    match,
    ...props
}) => {
    const [thread, onThreadUpdate] = useThreadByNumberInRepository(props.repo.id, threadNumber)

    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem(
                thread !== LOADING && thread !== null && !isErrorLike(thread)
                    ? { text: `#${thread.number}`, to: thread.url }
                    : undefined
            )
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [thread, setBreadcrumbItem])

    const onThreadDelete = useCallback(() => {
        if (thread !== LOADING && thread !== null && !isErrorLike(thread)) {
            props.history.push(`${thread.repository.url}/-/threads`)
        }
    }, [thread, props.history])

    const sidebarSections = useMemo<InfoSidebarSection[]>(
        () =>
            thread !== LOADING && thread !== null && !isErrorLike(thread)
                ? [
                      ...threadSidebarSections({
                          thread,
                          onThreadUpdate,
                          extensionsController: props.extensionsController,
                      }),
                      {
                          expanded: (
                              <ThreadDeleteButton
                                  {...props}
                                  thread={thread}
                                  buttonClassName="btn-link"
                                  className="btn-sm px-0 text-decoration-none"
                                  onDelete={onThreadDelete}
                              />
                          ),
                      },
                  ]
                : [],
        [thread, onThreadDelete, onThreadUpdate, props]
    )

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
            <style>{`.user-area-header, .org-header { display: none; } .org-area > .container, .user-area > .container { margin: unset; margin-top: unset !important; width: 100%; max-width: unset !important; overflow:hidden; padding: unset; } /* TODO!(sqs): hack */`}</style>
            <WithSidebar
                sidebarPosition="right"
                sidebar={<InfoSidebar sections={sidebarSections} />}
                className="flex-1"
            >
                <OverviewPagesArea<ThreadAreaContext>
                    context={context}
                    header={header}
                    overviewComponent={ThreadOverview}
                    pages={[
                        {
                            title: 'Activity',
                            path: '/',
                            exact: true,
                            render: () => <ThreadActivity {...context} className={PAGE_CLASS_NAME} />,
                        },
                        {
                            title: 'Changes',
                            path: '/changes',
                            exact: true,
                            render: () => <ThreadFileDiffsList {...context} className={PAGE_CLASS_NAME} />,
                        },
                        {
                            title: 'Commits',
                            path: '/commits',
                            render: () => <ThreadCommitsList {...context} className={PAGE_CLASS_NAME} />,
                        },
                        {
                            title: 'Diagnostics',
                            path: '/diagnostics',
                            render: () => <ThreadDiagnosticsList {...context} className={PAGE_CLASS_NAME} />,
                        },
                    ]}
                    location={props.location}
                    match={match}
                />
            </WithSidebar>
        </>
    )
}
