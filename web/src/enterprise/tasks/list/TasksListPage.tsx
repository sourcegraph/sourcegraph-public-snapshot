import React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { useDiagnostics } from '../../checks/detail/diagnostics/useDiagnostics'
import { TasksAreaTitle } from '../components/TasksAreaTitle'
import { TasksAreaContext } from '../global/TasksArea'
import { DiagnosticsList } from './DiagnosticsList'

interface Props extends TasksAreaContext, RouteComponentProps<{}> {}

/**
 * The tasks list page.
 */
export const TasksListPage: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const diagnosticsOrError = useDiagnostics(props.extensionsController)
    return (
        <div className="w-100 mt-3">
            <PageTitle title="Tasks" />
            <div className="container-fluid">
                <TasksAreaTitle />
            </div>
            <DiagnosticsList {...props} diagnosticsOrError={diagnosticsOrError} itemClassName="container-fluid" />
        </div>
    )
}
