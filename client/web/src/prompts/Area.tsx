import type { FunctionComponent } from 'react'

import { mdiPlus } from '@mdi/js'
import { Route, Routes } from 'react-router-dom'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, Icon, Link, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUserOnly } from '../auth/withAuthenticatedUser'
import { NotFoundPage } from '../components/HeroPage'
import { PageRoutes } from '../routes.constants'

import { DetailPage } from './DetailPage'
import { EditPage } from './EditPage'
import { ListPage } from './ListPage'
import { NewForm } from './NewForm'
import { PromptPage } from './Page'

/** The prompt area. */
export const Area: FunctionComponent<
    {
        authenticatedUser: AuthenticatedUser | null
    } & TelemetryV2Props
> = ({ authenticatedUser, telemetryRecorder }) => (
    <Routes>
        <Route
            path=""
            element={
                <PromptPage
                    title="Prompt Library"
                    actions={
                        authenticatedUser && (
                            <Button to={`${PageRoutes.Prompts}/new`} variant="primary" as={Link}>
                                <Icon aria-hidden={true} svgPath={mdiPlus} /> New prompt
                            </Button>
                        )
                    }
                >
                    <ListPage telemetryRecorder={telemetryRecorder} />
                </PromptPage>
            }
        />
        <Route
            path="new"
            element={
                <AuthenticatedUserOnly authenticatedUser={authenticatedUser}>
                    <PromptPage title="New prompt" breadcrumbs={<PageHeader.Breadcrumb>New</PageHeader.Breadcrumb>}>
                        <NewForm telemetryRecorder={telemetryRecorder} />
                    </PromptPage>
                </AuthenticatedUserOnly>
            }
        />
        <Route
            path=":id/edit"
            element={
                <AuthenticatedUserOnly authenticatedUser={authenticatedUser}>
                    <EditPage telemetryRecorder={telemetryRecorder} />
                </AuthenticatedUserOnly>
            }
        />
        <Route path=":id" element={<DetailPage telemetryRecorder={telemetryRecorder} />} />
        <Route path="*" element={<NotFoundPage pageType="prompt" />} />
    </Routes>
)
