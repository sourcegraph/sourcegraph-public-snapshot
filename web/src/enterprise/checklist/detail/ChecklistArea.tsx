import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { from } from 'rxjs'
import { filter, first, map, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { useEffectAsync } from '../../../util/useEffectAsync'
import { getCodeActions, getDiagnosticInfos } from '../../threads/detail/backend'
import { Checklist } from '../checklist'
import { ChecklistsAreaContext } from '../global/ChecklistsArea'
import { ChecklistFilesPage } from './files/ChecklistFilesPage'
import { ChecklistAreaNavbar } from './navbar/ChecklistAreaNavbar'
import { ChecklistOverview } from './overview/ChecklistOverview'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends ChecklistsAreaContext, RouteComponentProps<{ checklistID: string }> {}

export interface ChecklistAreaContext {
    /** The checklist. */
    checklist: Checklist
}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single checklist.
 */
export const ChecklistArea: React.FunctionComponent<Props> = props => {
    const [checklistOrError, setChecklistOrError] = useState<typeof LOADING | Checklist | ErrorLike>(LOADING)

    useEffectAsync(async () => {
        try {
            // TODO!(sqs)
            setChecklistOrError(
                await getDiagnosticInfos(props.extensionsController)
                    .pipe(
                        filter(diagnostics => diagnostics.length > 0),
                        first(),
                        map(diagnostics => diagnostics[0]),
                        switchMap(diagnostic =>
                            getCodeActions({ diagnostic, extensionsController: props.extensionsController }).pipe(
                                filter(codeActions => codeActions.length > 0),
                                first(),
                                map(codeActions => ({ diagnostic, codeActions }))
                            )
                        )
                    )
                    .toPromise()
            )
        } catch (err) {
            setChecklistOrError(asError(err))
        }
    }, [props.extensionsController])
    if (checklistOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(checklistOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={checklistOrError.message} />
    }

    const context: ChecklistsAreaContext &
        ChecklistAreaContext & {
            areaURL: string
        } = {
        ...props,
        checklist: checklistOrError,
        areaURL: props.match.url,
    }

    return (
        <div className="checklist-area flex-1 d-flex overflow-hidden">
            <div className="d-flex flex-column flex-1 overflow-auto">
                <ErrorBoundary location={props.location}>
                    <ChecklistOverview
                        {...context}
                        location={props.location}
                        history={props.history}
                        className="container flex-0 pb-3"
                    />
                    <div className="w-100 border-bottom" />
                    <ChecklistAreaNavbar {...context} className="flex-0 sticky-top bg-body" />
                </ErrorBoundary>
                <ErrorBoundary location={props.location}>
                    <Switch>
                        <Route
                            path={props.match.url}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <p>TODO!(sqs) empty</p>}
                        />
                        <Route
                            path={`${props.match.url}/files`}
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => <ChecklistFilesPage {...context} {...routeComponentProps} />}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}
