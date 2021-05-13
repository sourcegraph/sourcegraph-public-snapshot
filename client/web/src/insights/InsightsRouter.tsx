import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'

import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { AuthenticatedUser } from '../auth'
import { HeroPage } from '../components/HeroPage'
import { lazyComponent } from '../util/lazyComponent'

import { SearchInsightCreationPageProps } from './pages/creation/search-insight/SearchInsightCreationPage'
import { InsightsPageProps } from './pages/dashboard/InsightsPage'
import { EditInsightPageProps } from './pages/edit/EditInsightPage'

const InsightsLazyPage = lazyComponent(() => import('./pages/dashboard/InsightsPage'), 'InsightsPage')

const IntroCreationLazyPage = lazyComponent(
    () => import('./pages/creation/intro/IntroCreationPage'),
    'IntroCreationPage'
)

const SearchInsightCreationLazyPage = lazyComponent(
    () => import('./pages/creation/search-insight/SearchInsightCreationPage'),
    'SearchInsightCreationPage'
)

const LangStatsInsightCreationLazyPage = lazyComponent(
    () => import('./pages/creation/lang-stats/LangStatsInsightCreationPage'),
    'LangStatsInsightCreationPage'
)

const EditInsightLazyPage = lazyComponent(() => import('./pages/edit/EditInsightPage'), 'EditInsightPage')

/**
 * Feature flag for new code insights creation UI.
 * */
const CREATION_UI_ENABLED_KEY = 'enableCodeInsightCreationUI'

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * sub-components withing app tree.
 */
export interface InsightsRouterProps
    extends RouteComponentProps,
        Omit<InsightsPageProps, 'isCreationUIEnabled'>,
        SearchInsightCreationPageProps,
        EditInsightPageProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     * */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations' | 'username'> | null
}

/** Main Insight routing component. Main entry point to code insights UI. */
export const InsightsRouter: React.FunctionComponent<InsightsRouterProps> = props => {
    const { match, ...outerProps } = props
    const [isCreationUIEnabled] = useLocalStorage(CREATION_UI_ENABLED_KEY, false)

    return (
        <Switch>
            <Route
                render={props => (
                    <InsightsLazyPage isCreationUIEnabled={isCreationUIEnabled} {...outerProps} {...props} />
                )}
                path={match.url}
                exact={true}
            />

            {isCreationUIEnabled && (
                <>
                    <Route
                        path={`${match.url}/create-search-insight`}
                        render={props => <SearchInsightCreationLazyPage {...outerProps} {...props} />}
                    />

                    <Route
                        path={`${match.url}/create-lang-stats-insight`}
                        render={props => <LangStatsInsightCreationLazyPage {...outerProps} {...props} />}
                    />

                    <Route path={`${match.url}/create-intro`} component={IntroCreationLazyPage} />

                    <Route
                        path={`${match.url}/edit/:insightID`}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        render={(props: RouteComponentProps<{ insightID: string }>) => (
                            <EditInsightLazyPage
                                platformContext={outerProps.platformContext}
                                authenticatedUser={outerProps.authenticatedUser}
                                settingsCascade={outerProps.settingsCascade}
                                insightID={props.match.params.insightID}
                            />
                        )}
                    />
                </>
            )}

            <Route component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    )
}
