import React, { useCallback, useEffect, useState } from 'react'

import { mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'

import { CodyPageIcon } from './CodyPageIcon'

import styles from './CodyPage.module.scss'

interface CodePageProps {
    authenticatedUser: AuthenticatedUser | null
    telemetryService: TelemetryService
}

export const CodyPage: React.FunctionComponent<CodePageProps> = ({ authenticatedUser }) => {
    return (
        <Page>
            <PageTitle title="Cody AI" />
            <PageHeader
                actions={
                    <Button variant="primary" disabled={true}>
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> New chat
                    </Button>
                }
                description={
                    <>
                        Cody answers code questions and writes code for you by reading your entire codebase and the code
                        graph.
                    </>
                }
                className="mb-3"
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={CodyPageIcon}>Cody AI</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>
        </Page>
    )
}
