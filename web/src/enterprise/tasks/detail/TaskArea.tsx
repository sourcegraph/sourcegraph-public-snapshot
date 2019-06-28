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
import { TasksAreaContext } from '../global/TasksArea'
import { Task } from '../task'
import { TaskFilesPage } from './files/TaskFilesPage'
import { TaskAreaNavbar } from './navbar/TaskAreaNavbar'
import { TaskOverview } from './overview/TaskOverview'

const NotFoundPage = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested page was not found." />
)

interface Props extends TasksAreaContext, RouteComponentProps<{ taskID: string }> {}

export interface TaskAreaContext {
    /** The task. */
    task: Task
}

const LOADING: 'loading' = 'loading'

/**
 * The area for a single task.
 */
export const TaskArea: React.FunctionComponent<Props> = props => {
    const [taskOrError, setTaskOrError] = useState<typeof LOADING | Task | ErrorLike>(LOADING)

    useEffectAsync(async () => {
        try {
            // TODO!(sqs)
            setTaskOrError(
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
            setTaskOrError(asError(err))
        }
    }, [props.extensionsController])
    if (taskOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(taskOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={taskOrError.message} />
    }

    const context: TasksAreaContext &
        TaskAreaContext & {
            areaURL: string
        } = {
        ...props,
        task: taskOrError,
        areaURL: props.match.url,
    }

    return (
        <div className="task-area flex-1 d-flex overflow-hidden">
            <div className="d-flex flex-column flex-1 overflow-auto">
                <ErrorBoundary location={props.location}>
                    <TaskOverview
                        {...context}
                        location={props.location}
                        history={props.history}
                        className="container flex-0 pb-3"
                    />
                    <div className="w-100 border-bottom" />
                    <TaskAreaNavbar {...context} className="flex-0 sticky-top bg-body" />
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
                            render={routeComponentProps => <TaskFilesPage {...context} {...routeComponentProps} />}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                </ErrorBoundary>
            </div>
        </div>
    )
}
