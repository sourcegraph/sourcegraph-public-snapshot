import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'

import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { HeroPage } from '../components/HeroPage'
import { lazyComponent } from '../util/lazyComponent'

import { SearchInsightCreationPageProps } from './pages/creation/search-insight/SearchInsightCreationPage'
import { InsightsPageProps } from './pages/dashboard/InsightsPage'

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
        SearchInsightCreationPageProps {}

/** Main Insight routing component. Main entry point to code insights UI. */
export const InsightsRouter: React.FunctionComponent<InsightsRouterProps> = props => {
    const { match, ...outerProps } = props
    const [isCreationUIEnabled] = useLocalStorage(CREATION_UI_ENABLED_KEY, false)

    return (
        <Switch>
            <Route
                /* eslint-disable-next-line react/jsx-no-bind */
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
                        /* eslint-disable-next-line react/jsx-no-bind */
                        render={props => <SearchInsightCreationLazyPage {...outerProps} {...props} />}
                    />

                    <Route
                        path={`${match.url}/create-lang-stats-insight`}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        render={props => <LangStatsInsightCreationLazyPage {...outerProps} {...props} />}
                    />

                    <Route path={`${match.url}/create-intro`} component={IntroCreationLazyPage} />
                </>
            )}

            <Route component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    )
}
