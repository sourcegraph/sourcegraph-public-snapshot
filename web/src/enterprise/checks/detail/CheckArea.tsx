import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useEffect, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { of } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { Check, queryCheck } from '../data'
import { ChecksAreaContext } from '../global/ChecksArea'
import { CheckActivityPage } from './CheckActivityPage'
import { CheckAreaHeader } from './CheckAreaHeader'
import { CheckManagePage } from './CheckManagePage'
import { CheckOverviewPage } from './CheckOverviewPage'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested check page was not found." />
)

interface Props extends ChecksAreaContext, RouteComponentProps<{ checkID: string }> {}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single check.
 */
export const CheckArea: React.FunctionComponent<Props> = props => {
    const [checkOrError, setCheckOrError] = useState<typeof LOADING | Check | ErrorLike>(LOADING)

    useEffect(
        () => {
            const subscription = of(props.match.params.checkID)
                .pipe(switchMap(checkID => queryCheck(checkID)))
                .subscribe(checkOrError => setCheckOrError(checkOrError))
            return () => subscription.unsubscribe()
        },
        [props.match.params.checkID]
    )

    if (checkOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(checkOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={checkOrError.message} />
    }

    const context: ChecksAreaContext & { check: Check; areaURL: string } = {
        ...props,
        check: checkOrError,
        areaURL: props.match.url,
    }

    return (
        <div className="check-area area--vertical">
            <CheckAreaHeader {...context} />
            <div className="container pt-3">
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route
                            path={props.match.url}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <CheckOverviewPage {...routeComponentProps} {...context} />}
                        />
                        <Route
                            path={`${props.match.url}/activity`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <CheckActivityPage {...routeComponentProps} {...context} />}
                        />
                        <Route
                            path={`${props.match.url}/manage`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <CheckManagePage {...routeComponentProps} {...context} />}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}
