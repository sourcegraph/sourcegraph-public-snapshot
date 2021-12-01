import React, { useEffect, useRef, useState } from 'react'
import { Route, Switch, useRouteMatch } from 'react-router'
import { NavLink } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { Button, Container, PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { Badge } from '../../../../../components/Badge'
import { FeedbackPromptContent } from '../../../../../nav/Feedback/FeedbackPrompt'
import { Popover } from '../../../../insights/components/popover/Popover'
import { CatalogEntityFiltersProps } from '../../../core/entity-filters'
import { EntityList } from '../components/entity-list/EntityList'
import { OverviewEntityGraph } from '../components/overview-content/OverviewEntityGraph'

import styles from './OverviewPage.module.scss'

interface Props extends CatalogEntityFiltersProps, TelemetryProps {}

/**
 * The catalog overview page.
 */
export const OverviewPage: React.FunctionComponent<Props> = ({ filters, onFiltersChange, telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogOverview')
    }, [telemetryService])

    const match = useRouteMatch()

    return (
        <Page>
            <PageHeader
                path={[{ icon: CatalogIcon, text: 'Catalog' }]}
                className="mb-4"
                description="Explore software components, services, libraries, APIs, and more."
                actions={<FeedbackPopoverButton />}
            />

            <ul className="nav nav-tabs w-100 mb-2">
                <li className="nav-item">
                    <NavLink to={match.url} exact={true} className="nav-link px-3">
                        List
                    </NavLink>
                </li>
                <li className="nav-item">
                    <NavLink to={`${match.url}/graph`} exact={true} className="nav-link px-3">
                        Graph
                    </NavLink>
                </li>
            </ul>

            <Switch>
                <Route path={match.url} exact={true}>
                    <Container className="p-0 mb-2 flex-grow-1">
                        <EntityList filters={filters} onFiltersChange={onFiltersChange} size="lg" />
                    </Container>
                </Route>
                <Route path={`${match.url}/graph`} exact={true}>
                    <OverviewEntityGraph />
                </Route>
            </Switch>
        </Page>
    )
}

const FeedbackPopoverButton: React.FunctionComponent = () => {
    const buttonReference = useRef<HTMLButtonElement>(null)
    const [isVisible, setVisibility] = useState(false)

    return (
        <div className="d-flex align-items-center px-2">
            <Badge status="wip" className="text-uppercase mr-2" />
            <Button ref={buttonReference} variant="link" size="sm">
                Share feedback
            </Button>
            <Popover
                isOpen={isVisible}
                target={buttonReference}
                onVisibilityChange={setVisibility}
                className={styles.feedbackPrompt}
            >
                <FeedbackPromptContent
                    closePrompt={() => setVisibility(false)}
                    textPrefix="Catalog: "
                    routeMatch="/catalog"
                />
            </Popover>
        </div>
    )
}
