import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useCallback } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap, switchMapTo } from 'rxjs/operators'
import { ConfiguredRegistryExtension, toConfiguredRegistryExtension } from '../../../../shared/src/extensions/extension'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { createAggregateError, ErrorLike, isErrorLike, asError } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { RouteDescriptor } from '../../util/contributions'
import { ExtensionsAreaRouteContext } from '../ExtensionsArea'
import { ExtensionAreaHeader, ExtensionAreaHeaderNavItem } from './ExtensionAreaHeader'
import { ThemeProps } from '../../../../shared/src/theme'
import { ErrorMessage } from '../../components/alerts'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { useEventObservable } from '../../../../shared/src/util/useObservable'
import { ExtensionsAreaHeaderActionButton } from '../ExtensionsAreaHeader'

export const registryExtensionFragment = gql`
    fragment RegistryExtensionFields on RegistryExtension {
        id
        publisher {
            __typename
            ... on User {
                id
                username
                displayName
                url
            }
            ... on Org {
                id
                name
                displayName
                url
            }
        }
        extensionID
        extensionIDWithoutRegistry
        name
        manifest {
            raw
            description
        }
        createdAt
        updatedAt
        publishedAt
        url
        remoteURL
        registryName
        isLocal
        isWorkInProgress
        viewerCanAdminister
    }
`

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface ExtensionAreaRoute extends RouteDescriptor<ExtensionAreaRouteContext> {}

export interface ExtensionAreaProps
    extends ExtensionsAreaRouteContext,
        RouteComponentProps<{ extensionID: string }>,
        ThemeProps,
        TelemetryProps {
    routes: readonly ExtensionAreaRoute[]
    extensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[]
}

/**
 * Properties passed to all page components in the registry extension area.
 */
export interface ExtensionAreaRouteContext
    extends SettingsCascadeProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps {
    /** The extension registry area main URL. */
    url: string

    /** The extension that is the subject of the page. */
    extension: ConfiguredRegistryExtension<GQL.IRegistryExtension>

    onDidUpdateExtension: () => void

    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null
}

/**
 * An extension area.
 */
export const ExtensionArea: React.FunctionComponent<ExtensionAreaProps> = props => {
    const [onDidUpdateExtension, extensionOrError] = useEventObservable(
        useCallback(
            (refreshRequests: Observable<void>) =>
                refreshRequests.pipe(
                    startWith<void>(undefined),
                    switchMapTo(queryExtension(props.match.params.extensionID)),
                    catchError((error): [ErrorLike] => [asError(error)])
                ),
            [props.match.params.extensionID]
        )
    )

    if (!extensionOrError) {
        return null // loading
    }
    if (isErrorLike(extensionOrError)) {
        return (
            <HeroPage
                icon={AlertCircleIcon}
                title="Error"
                subtitle={<ErrorMessage error={extensionOrError} history={props.history} />}
            />
        )
    }

    // The URL, without the trailing "/-" that `props.match.url` includes on sub-pages.
    const url = props.match.url.replace(/\/-\/?$/, '')

    const context: ExtensionAreaRouteContext = {
        url,
        authenticatedUser: props.authenticatedUser,
        onDidUpdateExtension,
        settingsCascade: props.settingsCascade,
        extension: extensionOrError,
        platformContext: props.platformContext,
        isLightTheme: props.isLightTheme,
        telemetryService: props.telemetryService,
    }

    return (
        <div className="registry-extension-area">
            <ExtensionAreaHeader {...props} {...context} navItems={props.extensionAreaHeaderNavItems} />
            <div className="container pt-3">
                <ErrorBoundary location={props.location}>
                    <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                        <Switch>
                            {props.routes.map(
                                /* eslint-disable react/jsx-no-bind */
                                ({ path, render, exact, condition = () => true }) =>
                                    condition(context) && (
                                        <Route
                                            path={url + path}
                                            exact={exact}
                                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                            render={routeComponentProps =>
                                                render({ ...context, ...routeComponentProps })
                                            }
                                        />
                                    )
                                /* eslint-enable react/jsx-no-bind */
                            )}
                            <Route key="hardcoded-key" component={NotFoundPage} />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}

function queryExtension(extensionID: string): Observable<ConfiguredRegistryExtension<GQL.IRegistryExtension>> {
    return queryGraphQL(
        gql`
            query RegistryExtension($extensionID: String!) {
                extensionRegistry {
                    extension(extensionID: $extensionID) {
                        ...RegistryExtensionFields
                    }
                }
            }
            ${registryExtensionFragment}
        `,
        { extensionID }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.extensionRegistry || !data.extensionRegistry.extension) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.extension
        }),
        map(registryExtension => toConfiguredRegistryExtension(registryExtension))
    )
}
