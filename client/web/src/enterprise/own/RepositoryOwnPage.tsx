import React, { useCallback, useEffect, useState } from 'react'

import { mdiAccount, mdiPencil, mdiPlus } from '@mdi/js'
import { Navigate } from 'react-router-dom'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import {
    Button,
    H1,
    Icon,
    Link,
    LoadingSpinner,
    PageHeader,
    ProductStatusBadge,
    ButtonLink,
} from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { TreeOwnershipPanel } from '../../repo/blob/own/TreeOwnershipPanel'

import { AddOwnerModal } from './AddOwnerModal'
import { RepositoryOwnAreaPageProps } from './RepositoryOwnEditPage'

import styles from './RepositoryOwnPageContents.module.scss'

const BREADCRUMB = { key: 'own', element: 'Ownership' }

export const RepositoryOwnPage: React.FunctionComponent<RepositoryOwnAreaPageProps> = ({
    useBreadcrumb,
    repo,
    telemetryService,
}) => {
    const queryParameters = new URLSearchParams(location.search)
    const path = queryParameters.get('path') ?? ''

    useBreadcrumb(BREADCRUMB)

    const [ownEnabled, status] = useFeatureFlag('search-ownership')
    const [openAddOwnerModal, setOpenAddOwnerModal] = useState<boolean>(false)
    const onClickAdd = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenAddOwnerModal(true)
    }, [])
    const closeModal = useCallback(() => {
        setOpenAddOwnerModal(false)
    }, [])

    useEffect(() => {
        if (status !== 'initial' && ownEnabled) {
            telemetryService.log('repoPage:ownershipPage:viewed')
        }
    }, [status, ownEnabled, telemetryService])

    if (status === 'initial') {
        return (
            <div className="container d-flex justify-content-center mt-3">
                <LoadingSpinner /> Loading...
            </div>
        )
    }

    if (!ownEnabled) {
        return <Navigate to={repo.url} replace={true} />
    }

    return (
        <>
            <Page>
                <PageTitle title={`Ownership for ${displayRepoName(repo.name)}`} />
                <div className={styles.actionButtons}>
                    <ButtonLink
                        aria-label="Navigate to edit ownership page"
                        className="mr-2"
                        variant="secondary"
                        to={`${repo.url}/-/own/edit`}
                    >
                        <Icon aria-hidden={true} svgPath={mdiPencil} /> Edit Ownership
                    </ButtonLink>
                    <Button aria-label="Add an owner" variant="success" onClick={onClickAdd}>
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add owner
                    </Button>
                </div>

                <PageHeader
                    description={
                        <>
                            Sourcegraph Own can provide code ownership data for this repository via an upload or a
                            committed CODEOWNERS file. <Link to="/help/own">Learn more about Sourcegraph Own.</Link>
                        </>
                    }
                >
                    <H1 as="h2" className="d-flex align-items-center">
                        <Icon svgPath={mdiAccount} aria-hidden={true} />
                        <span className="ml-2">Ownership</span>
                        <ProductStatusBadge status="beta" className="ml-2" />
                    </H1>
                </PageHeader>

                <TreeOwnershipPanel repoID={repo.id} filePath={path} telemetryService={telemetryService} />
            </Page>
            {openAddOwnerModal && <AddOwnerModal repoID={repo.id} path={path} onCancel={closeModal} />}
        </>
    )
}
