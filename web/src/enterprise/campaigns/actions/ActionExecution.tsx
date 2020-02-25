import * as React from 'react'
import * as H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { ThemeProps } from '../../../../../shared/src/theme'
import classNames from 'classnames'
import { MonacoSettingsEditor } from '../../../settings/MonacoSettingsEditor'
import { ActionJob } from './ActionJob'
import SyncIcon from 'mdi-react/SyncIcon'
import { useObservable } from '../../../util/useObservable'
import { fetchActionExecutionByID } from './backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { interval, merge, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { Link } from '../../../../../shared/src/components/Link'

interface Props extends ThemeProps {
    actionExecutionID: string

    location: H.Location
    history: H.History
}

export const ActionExecution: React.FunctionComponent<Props> = ({
    actionExecutionID,
    isLightTheme,
    history,
    location,
}) => {
    const execution = useObservable(
        React.useMemo(
            () =>
                merge(of(undefined), interval(2000)).pipe(switchMap(() => fetchActionExecutionByID(actionExecutionID))),
            [actionExecutionID]
        )
    )
    if (execution === undefined) {
        return <LoadingSpinner />
    }
    if (execution === null) {
        return <h3>Execution not found!</h3>
    }
    // const [execution] = React.useState<GQL.IActionExecution>({
    //     __typename: 'ActionExecution',
    //     id: 'asd',
    //     invokationReason: GQL.ActionExecutionInvokationReason.MANUAL,
    //     env: [{ __typename: 'ActionEnvVar', key: 'FORCE_COLOR', value: '1' }],
    //     actionDefinition: '',
    //     actionWorkspace: null,
    //     campaignPlan: null,
    //     // TODO: this is ugly
    //     action: {
    //         __typename: 'Action',
    //         id: 'asd',
    //         actionDefinition,
    //         actionWorkspace: null,
    //         schedule: '30 */2 * * *',
    //         cancelPreviousScheduledExecution: false,
    //         env: [{ __typename: 'ActionEnvVar', key: 'SG_SUPER_KEY', value: '<REDACTED>' }],
    //         savedSearch: {
    //             description: 'Repo has core.js <3.0.0 installed',
    //         } as any,
    //         campaign: null,
    //         actionExecutions: {
    //             __typename: 'ActionExecutionConnection',
    //             totalCount: 0,
    //             nodes: [] as GQL.IActionExecution[],
    //         },
    //     },
    //     state: {
    //         __typename: 'BackgroundProcessStatus',
    //         completedCount: 10,
    //         errors: [],
    //         pendingCount: 1500,
    //         state: GQL.BackgroundProcessState.PROCESSING,
    //     },
    //     jobs: {
    //         __typename: 'ActionJobConnection',
    //         totalCount: 10,
    //         nodes: [
    //             {
    //                 __typename: 'ActionJob',
    //                 id: 'asd',
    //                 image: 'alpine',
    //                 command: null,
    //                 baseRevision: 'master',
    //                 diff: null,
    //                 log: null,
    //                 executionStart: new Date().toISOString(),
    //                 executionEnd: undefined,
    //                 runner: {
    //                     __typename: 'Runner',
    //                     id: 'asda',
    //                     name: 'sg-dev1',
    //                     description: 'macOS 10.15.2',
    //                 } as any,
    //                 repository: {
    //                     name: 'github.com/sourcegraph/sourcegraph',
    //                 } as any,
    //                 state: GQL.ActionJobState.RUNNING,
    //             },
    //             {
    //                 __typename: 'ActionJob',
    //                 id: 'asd',
    //                 image: 'alpine',
    //                 command: null,
    //                 baseRevision: 'master',
    //                 diff: null,
    //                 log: null,
    //                 executionStart: undefined,
    //                 executionEnd: undefined,
    //                 runner: null,
    //                 repository: {
    //                     name: 'github.com/sourcegraph/about',
    //                 } as any,
    //                 state: GQL.ActionJobState.PENDING,
    //             },
    //             {
    //                 __typename: 'ActionJob',
    //                 id: 'asd',
    //                 image: 'alpine',
    //                 command: null,
    //                 baseRevision: 'master',
    //                 executionStart: subHours(new Date(), 1).toISOString(),
    //                 executionEnd: subMinutes(new Date(), 30).toISOString(),
    //                 runner: {
    //                     id: 'asda',
    //                     name: 'sg-dev1',
    //                     description: 'macOS 10.15.2',
    //                 } as any,
    //                 repository: {
    //                     name: 'github.com/sourcegraph/src-cli',
    //                 } as any,
    //                 diff: {
    //                     fileDiffs: {
    //                         diffStat: {
    //                             __typename: 'DiffStat',
    //                             added: 10,
    //                             changed: 20,
    //                             deleted: 10,
    //                         },
    //                     },
    //                 } as any,
    //                 state: GQL.ActionJobState.COMPLETED,
    //                 log: actionLog,
    //             },
    //             {
    //                 __typename: 'ActionJob',
    //                 id: 'asd',
    //                 image: 'alpine',
    //                 command: null,
    //                 baseRevision: 'master',
    //                 diff: null,
    //                 executionStart: subHours(new Date(), 1).toISOString(),
    //                 executionEnd: subMinutes(new Date(), 10).toISOString(),
    //                 runner: {
    //                     id: 'asda',
    //                     name: 'sg-dev1',
    //                     description: 'macOS 10.15.2',
    //                 } as any,
    //                 repository: {
    //                     name: 'github.com/sourcegraph/javascript-typescript-langserver',
    //                 } as any,
    //                 state: GQL.ActionJobState.ERRORED,
    //                 log: actionLog,
    //             },
    //         ],
    //     },
    // })
    return (
        <>
            <PageTitle title="Action execution #19292" />
            <h1 className={classNames(isLightTheme && 'text-info')}>Running action #19292</h1>
            {execution.invokationReason === GQL.ActionExecutionInvokationReason.SCHEDULE && (
                <div className="alert alert-info">
                    This execution is part of a scheduled action.
                    <br />
                    <code>{execution.action.schedule}</code>
                </div>
            )}
            {execution.invokationReason === GQL.ActionExecutionInvokationReason.SAVED_SEARCH && (
                <div className="alert alert-info">
                    This execution is run because the results of saved search "
                    <a href="">
                        <i>{execution.action.savedSearch?.description}</i>
                    </a>
                    " changed.
                </div>
            )}
            <h2>Action definition</h2>
            <MonacoSettingsEditor
                isLightTheme={isLightTheme}
                readOnly={true}
                language="json"
                value={execution.definition.steps}
                height={200}
                className="mb-3"
            ></MonacoSettingsEditor>
            <h2>Action status</h2>
            <div>
                <div className="alert alert-info mt-1">
                    <h3>Want faster execution? To add more runners:</h3>
                    <p>Use the below token to register your runner to this Sourcegraph instance</p>
                    <input className="form-control mb-2" readOnly={true} value="KK3DK99AA1291S8" />
                    <div>
                        <code>
                            src runner register --sourcegraph-url={window.location.protocol}//{window.location.host}{' '}
                            --token=KK3DK99AA1291S8
                        </code>
                    </div>
                    or
                    <div>
                        <code>
                            kubectl create secret generic sg-token --from-literal=token=KK3DK99AA1291S8
                            --from-literal=sourcegraphUrl={window.location.protocol}//{window.location.host}
                            <br />
                            kubectl apply -f {window.location.protocol}//{window.location.host}
                            /.api/runner-kubeconfig.yml
                        </code>
                    </div>
                </div>
                <SyncIcon className="icon-inline" /> Action is running for 03h:15:12, estimated remaining time:
                10h:51:49
            </div>
            {execution.status.state === GQL.BackgroundProcessState.PROCESSING && (
                <div className="progress">
                    <div
                        className="progress-bar"
                        /* eslint-disable-next-line react/forbid-dom-props */
                        style={{
                            width:
                                (execution.status.pendingCount /
                                    (execution.status.pendingCount + execution.status.completedCount)) *
                                    100 +
                                '%',
                        }}
                    >
                        &nbsp;
                    </div>
                </div>
            )}
            <p className="text-center w-100">
                {execution.status.pendingCount} / {execution.status.pendingCount + execution.status.completedCount} jobs
                ran
            </p>
            <h2>Action jobs</h2>
            <ul className="list-group mb-3">
                {execution.jobs.nodes.map(actionJob => (
                    <ActionJob
                        isLightTheme={isLightTheme}
                        actionJob={actionJob}
                        key={actionJob.id}
                        history={history}
                        location={location}
                    />
                ))}
            </ul>
            <h2>Action summary</h2>
            <p>
                The action execution finished
                {execution.status.state === GQL.BackgroundProcessState.ERRORED
                    ? ', but some tasks failed'
                    : ' successfully'}
                . <span className="badge badge-secondary">{execution.jobs.nodes.filter(job => job.diff).length}</span>{' '}
                patches have been created.
            </p>
            {execution.campaignPlan ? (
                <div className="mb-3">
                    <Link to={`/campaigns/new?plan=${execution.campaignPlan.id}`} className="btn btn-primary">
                        Create campaign
                    </Link>
                    <Link to={`/campaigns/update?plan=${execution.campaignPlan.id}`} className="btn btn-primary ml-2">
                        Update existing campaign
                    </Link>
                </div>
            ) : (
                <p>
                    Please stand by as we generate the campaign plan. Creating and updating a campaign will be possible
                    soon.
                </p>
            )}
        </>
    )
}
