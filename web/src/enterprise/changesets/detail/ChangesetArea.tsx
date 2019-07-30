import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { RepositoryChangesetsAreaContext } from '../repository/RepositoryChangesetsArea'
import { ChangesetOverview } from './ChangesetOverview'
import { ChangesetCommitsList } from './commits/ChangesetCommitsList'
import { useChangesetByNumberInRepository } from './useChangesetByNumberInRepository'

export interface ChangesetAreaContext
    extends Pick<RepositoryChangesetsAreaContext, Exclude<keyof RepositoryChangesetsAreaContext, 'repository'>> {
    /** The changeset, queried from the GraphQL API. */
    changeset: GQL.IChangeset

    /** Called to refresh the changeset. */
    onChangesetUpdate: () => void

    location: H.Location
    history: H.History
}

interface Props
    extends Pick<ChangesetAreaContext, Exclude<keyof ChangesetAreaContext, 'changeset' | 'onChangesetUpdate'>>,
        RouteComponentProps<{}> {
    /**
     * The changeset ID in its repository (i.e., the `Changeset.number` GraphQL API field).
     */
    changesetNumber: GQL.IChangeset['number']

    header: React.ReactFragment
}

const LOADING = 'loading' as const

const PAGE_CLASS_NAME = 'container mt-4'

/**
 * The area for a single changeset.
 */
export const ChangesetArea: React.FunctionComponent<Props> = ({
    header,
    changesetNumber,
    setBreadcrumbItem,
    match,
    ...props
}) => {
    const [changeset, onChangesetUpdate] = useChangesetByNumberInRepository(props.repo.id, changesetNumber)

    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem(
                changeset !== LOADING && changeset !== null && !isErrorLike(changeset)
                    ? { text: `#${changeset.number}`, to: changeset.url }
                    : undefined
            )
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [changeset, setBreadcrumbItem])

    if (changeset === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (changeset === null) {
        return <HeroPage icon={AlertCircleIcon} title="Changeset not found" />
    }
    if (isErrorLike(changeset)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={changeset.message} />
    }

    const context: ChangesetAreaContext = {
        ...props,
        changeset,
        onChangesetUpdate,
        setBreadcrumbItem,
    }

    return (
        <>
            <style>{`.user-area-header, .org-header { display: none; } .org-area > .container, .user-area > .container { margin: unset; margin-top: unset !important; width: unset; padding: unset; } /* TODO!(sqs): hack */`}</style>
            <OverviewPagesArea<ChangesetAreaContext>
                context={context}
                header={header}
                overviewComponent={ChangesetOverview}
                pages={[
                    {
                        title: 'Commits',
                        path: '/commits',
                        render: () => <ChangesetCommitsList {...context} className={PAGE_CLASS_NAME} />,
                    },
                    // { title: 'Changes', path: '/changes', render: () => <ChangesetChangesListPage {...context} /> },
                ]}
                location={props.location}
                match={match}
            />
        </>
    )
}
