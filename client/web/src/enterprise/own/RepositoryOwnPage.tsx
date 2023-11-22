import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiAccount, mdiPencil, mdiPlus } from '@mdi/js'
import { useSearchParams } from 'react-router-dom'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { Button, H1, Icon, Link, PageHeader, ProductStatusBadge, ButtonLink } from '@sourcegraph/wildcard'

import { AddOwnerModal } from '../../components/own/AddOwnerModal'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { TreeOwnershipPanel } from '../../repo/blob/own/TreeOwnershipPanel'
import { FilePathBreadcrumbs } from '../../repo/FilePathBreadcrumbs'

import type { RepositoryOwnAreaPageProps } from './RepositoryOwnEditPage'

import styles from './RepositoryOwnPageContents.module.scss'

export const RepositoryOwnPage: React.FunctionComponent<RepositoryOwnAreaPageProps> = ({
    useBreadcrumb,
    repo,
    telemetryService,
}) => {
    const [searchParams] = useSearchParams()
    const filePath = searchParams.get('path') ?? ''

    useBreadcrumb(
        useMemo(() => {
            if (!filePath || !repo) {
                return
            }
            return {
                key: 'treePath',
                className: 'flex-shrink-past-contents',
                element: (
                    <FilePathBreadcrumbs
                        key="path"
                        repoName={repo.name}
                        revision="main"
                        filePath={filePath}
                        isDir={true}
                        telemetryService={telemetryService}
                    />
                ),
            }
        }, [filePath, repo, telemetryService])
    )

    useBreadcrumb({ key: 'own', element: 'Ownership' })

    const [openAddOwnerModal, setOpenAddOwnerModal] = useState<boolean>(false)
    const onClickAdd = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenAddOwnerModal(true)
    }, [])
    const closeModal = useCallback(() => {
        setOpenAddOwnerModal(false)
    }, [])

    useEffect(() => {
        telemetryService.log('repoPage:ownershipPage:viewed')
    }, [telemetryService])
    return (
        <>
            <Page>
                <PageTitle title={`Ownership for ${displayRepoName(repo.name)}`} />
                <div className={styles.actionButtons}>
                    <ButtonLink
                        aria-label="Navigate to upload CODEOWNERS page"
                        className="mr-2"
                        variant="secondary"
                        to={`${repo.url}/-/own/edit`}
                    >
                        <Icon aria-hidden={true} svgPath={mdiPencil} /> Upload CODEOWNERS
                    </ButtonLink>
                    <Button aria-label="Add an owner" variant="success" onClick={onClickAdd}>
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add owner
                    </Button>
                </div>

                <PageHeader
                    description={
                        <>
                            Code ownership data for this repository can be provided via an upload or a committed
                            CODEOWNERS file. <Link to="/help/own">Learn more about code ownership.</Link>
                        </>
                    }
                >
                    <H1 as="h2" className="d-flex align-items-center">
                        <Icon svgPath={mdiAccount} aria-hidden={true} />
                        <span className="ml-2">Ownership</span>
                        <ProductStatusBadge status="beta" className="ml-2" />
                    </H1>
                </PageHeader>

                <TreeOwnershipPanel repoID={repo.id} filePath={filePath} telemetryService={telemetryService} />
            </Page>
            {openAddOwnerModal && <AddOwnerModal repoID={repo.id} path={filePath} onCancel={closeModal} />}
        </>
    )
}
