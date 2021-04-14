import React from 'react';
import { RouteComponentProps, Switch, Route } from 'react-router';

import { lazyComponent } from '../util/lazyComponent';

import { InsightsPageProps } from './pages/InsightsPage';

const InsightsLazyPage = lazyComponent<InsightsPageProps, 'InsightsPage'>(
    () => import('./pages/InsightsPage'),
    'InsightsPage'
)

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * sub-components withing app tree.
 */
export interface InsightsRouterProps extends RouteComponentProps, InsightsPageProps {}

/** Main Insight routing component. Main entry point to code insights UI. */
export const InsightsRouter: React.FunctionComponent<InsightsRouterProps> = props => {
    const { match, ...outerProps } = props;

    return (
        <Switch>
            <Route
                /* eslint-disable-next-line react/jsx-no-bind */
                render={props => <InsightsLazyPage {...outerProps} {...props} />}
                path={match.url}
                exact={true}
            />

            <Route />
        </Switch>
    );
}
