import type { FunctionComponent, PropsWithChildren } from 'react'

import { mdiPlus } from '@mdi/js'
import { Route, Routes } from 'react-router-dom'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, Icon, Link, PageHeader } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { NotFoundPage } from '../components/HeroPage'

import { DetailPage } from './DetailPage'
import { EditPage } from './EditPage'
import { ListPage } from './ListPage'
import { NewForm } from './NewForm'
import { SavedSearchPage } from './Page'

interface Props extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

const AuthenticatedArea: FunctionComponent<PropsWithChildren<Props>> = ({ telemetryRecorder, isSourcegraphDotCom }) => (
    <Routes>
        <Route
            path=""
            element={
                <SavedSearchPage
                    title="Saved searches"
                    actions={
                        <Button to="new" variant="primary" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} /> New saved search
                        </Button>
                    }
                    data-testid="saved-searches-list-page"
                >
                    <ListPage telemetryRecorder={telemetryRecorder} />
                </SavedSearchPage>
            }
        />
        <Route
            path="new"
            element={
                <SavedSearchPage
                    title="New saved search"
                    breadcrumbs={<PageHeader.Breadcrumb>New</PageHeader.Breadcrumb>}
                >
                    <NewForm isSourcegraphDotCom={isSourcegraphDotCom} telemetryRecorder={telemetryRecorder} />
                </SavedSearchPage>
            }
        />
        <Route
            path=":id/edit"
            element={<EditPage isSourcegraphDotCom={isSourcegraphDotCom} telemetryRecorder={telemetryRecorder} />}
        />
        <Route path=":id" element={<DetailPage telemetryRecorder={telemetryRecorder} />} />
        <Route path="*" element={<NotFoundPage pageType="saved search" />} />
    </Routes>
)

/** The saved search area. */
export const Area = withAuthenticatedUser(AuthenticatedArea)
