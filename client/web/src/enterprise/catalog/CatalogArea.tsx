import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Switch, Route, useRouteMatch, RouteComponentProps } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { HeroPage } from '../../components/HeroPage'

import styles from './CatalogArea.module.scss'
import { useCatalogEntityFilters } from './core/entity-filters'
import { EntityDetailPage } from './pages/entity-detail/global/EntityDetailPage'
import { OverviewPage } from './pages/overview/global/OverviewPage'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {}

/**
 * The main entrypoint to the catalog UI.
 */
export const CatalogArea: React.FunctionComponent<Props> = ({ telemetryService, ...props }) => {
    const match = useRouteMatch()

    const { filters, onFiltersChange } = useCatalogEntityFilters()

    return (
        <div className={styles.container}>
            <Switch>
                <Route path={`${match.url}/entities/:name`}>
                    {(matchProps: RouteComponentProps<{ name: string }>) => (
                        <EntityDetailPage
                            key={1}
                            {...props}
                            entityName={matchProps.match.params.name}
                            filters={filters}
                            onFiltersChange={onFiltersChange}
                            telemetryService={telemetryService}
                        />
                    )}
                </Route>
                <Route path={match.url}>
                    <OverviewPage
                        filters={filters}
                        onFiltersChange={onFiltersChange}
                        telemetryService={telemetryService}
                    />
                </Route>
                <Route>
                    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
                </Route>
            </Switch>
        </div>
    )
}
