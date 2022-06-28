import React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

import { AuthenticatedUser } from '../../../auth'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { HeroPage } from '../../../components/HeroPage'
import { ExtensionsAreaRouteContext } from '../../../extensions/ExtensionsArea'

import { RegistryNewExtensionPage } from './RegistryNewExtensionPage'

import styles from './RegistryArea.module.scss'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)

interface Props extends RouteComponentProps<{}>, ExtensionsAreaRouteContext {}

/**
 * Properties passed to all page components in the registry area.
 */
export interface RegistryAreaPageProps extends PlatformContextProps, BreadcrumbSetters {
    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser | null
}

/**
 * The extension registry area.
 */
export const RegistryArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    platformContext,
    useBreadcrumb,
    setBreadcrumb,
    match,
}) => {
    const transferProps: RegistryAreaPageProps = {
        authenticatedUser,
        platformContext,
        useBreadcrumb,
        setBreadcrumb,
    }

    return (
        <div className={styles.registryArea}>
            <Switch>
                <Route
                    path={`${match.url}/new`}
                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                    exact={true}
                    render={routeComponentProps => (
                        <RegistryNewExtensionPage {...routeComponentProps} {...transferProps} />
                    )}
                />

                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </div>
    )
}
