import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { ThemeProps } from '../../../../../shared/src/theme'
import classNames from 'classnames'
import { MonacoSettingsEditor } from '../../../settings/MonacoSettingsEditor'
import { ActionJob } from './ActionJob'
import { subMinutes, subHours } from 'date-fns'
import SyncIcon from 'mdi-react/SyncIcon'

/*
TODO: the current model does not support:
- rerunning a job and retaining the old one in the list of jobs
- running multiple jobs per repo (monorepo support)

other TODOs:

- What happens with the "command" step type? Ideally, it would still be wrapped in a docker container, `ubuntu` base image for example.
    Otherwise, it has a security risk because runs can modify other run's environment to, for example, expose data.
    Also, weird and hard to debug errors may arise, as there is no guarantee that the environment is pristine after a run.
*/

interface Props extends ThemeProps {}

const actionDefinition = `{
    "scopeQuery": "repo:github.com/sd9/[b-z]* repohasfile:^package.json",
    "steps": [
        {
        "type": "command",
        "args": ["sed", "-i", "", "s/\\"main\\"/\\"es2015\\"/", "package.json"]
        }
    ]
}`

// IS_TTY=1 COLOR=1 yarn | npx ansi-to-html
const actionLog = `G<b>yarn install v1.21.0<span style="font-weight:normal;text-decoration:none;font-style:normal">
</span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal">G</span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal">[1/5]<span style="font-weight:normal;text-decoration:none;font-style:normal"> Validating package.json...
</span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal">G</span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal">[2/5]<span style="font-weight:normal;text-decoration:none;font-style:normal"> Resolving packages...
</span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal">G<span style="color:#0A0">success<span style="color:#FFF"> Already up-to-date.
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF">G</span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF">$ gulp generate<span style="font-weight:normal;text-decoration:none;font-style:normal">
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">[22:21:56] </span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">Using gulpfile ~/Code/sourcegraph/sourcegraph/gulpfile.js
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">[22:21:56] Starting 'generate'...
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">[22:21:56] Starting 'schema'...
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">[22:21:56] </span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">Starting 'graphQLTypes'...
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">[22:21:57] </span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">Finished 'schema' after 718 ms
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">[22:21:57] </span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">Finished 'graphQLTypes' after 1.06 s
[22:21:57] Finished 'generate' after 1.06 s
</span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal"></span></span></span></span></span></b><b><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="font-weight:normal;text-decoration:none;font-style:normal"><span style="color:#FFF"><span style="font-weight:normal;text-decoration:none;font-style:normal">GDone in 3.44s.
</span></span></span></span></span></b>`

export const ActionExecution: React.FunctionComponent<Props> = ({ isLightTheme }) => {
    const [execution] = React.useState<GQL.IActionExecution>({
        __typename: 'ActionExecution',
        id: 'asd',
        invokationReason: GQL.ActionExecutionInvokationReason.MANUAL,
        env: [{ __typename: 'ActionEnvVar', key: 'FORCE_COLOR', value: '1' }],
        actionDefinition: '',
        actionWorkspace: null,
        campaignPlan: null,
        // TODO: this is ugly
        action: {
            __typename: 'Action',
            id: 'asd',
            actionDefinition,
            actionWorkspace: null,
            schedule: '30 */2 * * *',
            cancelPreviousScheduledExecution: false,
            env: [{ __typename: 'ActionEnvVar', key: 'SG_SUPER_KEY', value: '<REDACTED>' }],
            savedSearch: {
                description: 'Repo has core.js <3.0.0 installed',
            } as any,
            campaign: null,
            actionExecutions: {
                __typename: 'ActionExecutionConnection',
                totalCount: 0,
                nodes: [] as GQL.IActionExecution[],
            },
        },
        state: {
            __typename: 'BackgroundProcessStatus',
            completedCount: 10,
            errors: [],
            pendingCount: 1500,
            state: GQL.BackgroundProcessState.PROCESSING,
        },
        jobs: {
            __typename: 'ActionJobConnection',
            totalCount: 10,
            nodes: [
                {
                    __typename: 'ActionJob',
                    id: 'asd',
                    image: 'alpine',
                    command: null,
                    baseRevision: 'master',
                    diff: null,
                    log: null,
                    executionStart: new Date().toISOString(),
                    executionEnd: undefined,
                    runner: {
                        __typename: 'Runner',
                        id: 'asda',
                        name: 'sg-dev1',
                        description: 'macOS 10.15.2',
                    } as any,
                    repository: {
                        name: 'github.com/sourcegraph/sourcegraph',
                    } as any,
                    state: GQL.ActionJobState.RUNNING,
                },
                {
                    __typename: 'ActionJob',
                    id: 'asd',
                    image: 'alpine',
                    command: null,
                    baseRevision: 'master',
                    diff: null,
                    log: null,
                    executionStart: undefined,
                    executionEnd: undefined,
                    runner: null,
                    repository: {
                        name: 'github.com/sourcegraph/about',
                    } as any,
                    state: GQL.ActionJobState.PENDING,
                },
                {
                    __typename: 'ActionJob',
                    id: 'asd',
                    image: 'alpine',
                    command: null,
                    baseRevision: 'master',
                    executionStart: subHours(new Date(), 1).toISOString(),
                    executionEnd: subMinutes(new Date(), 30).toISOString(),
                    runner: {
                        id: 'asda',
                        name: 'sg-dev1',
                        description: 'macOS 10.15.2',
                    } as any,
                    repository: {
                        name: 'github.com/sourcegraph/src-cli',
                    } as any,
                    diff: {
                        fileDiffs: {
                            diffStat: {
                                __typename: 'DiffStat',
                                added: 10,
                                changed: 20,
                                deleted: 10,
                            },
                        },
                    } as any,
                    state: GQL.ActionJobState.COMPLETED,
                    log: actionLog,
                },
                {
                    __typename: 'ActionJob',
                    id: 'asd',
                    image: 'alpine',
                    command: null,
                    baseRevision: 'master',
                    diff: null,
                    executionStart: subHours(new Date(), 1).toISOString(),
                    executionEnd: subMinutes(new Date(), 10).toISOString(),
                    runner: {
                        id: 'asda',
                        name: 'sg-dev1',
                        description: 'macOS 10.15.2',
                    } as any,
                    repository: {
                        name: 'github.com/sourcegraph/javascript-typescript-langserver',
                    } as any,
                    state: GQL.ActionJobState.ERRORED,
                    log: actionLog,
                },
            ],
        },
    })
    return (
        <>
            <PageTitle title="Action execution #19292" />
            <h1 className={classNames(isLightTheme && 'text-info')}>Running action #19292</h1>
            {execution.action.schedule && (
                <div className="alert alert-info">
                    This execution is part of a scheduled action.
                    <br />
                    <code>{execution.action.schedule}</code>
                </div>
            )}
            {execution.action.savedSearch && (
                <div className="alert alert-info">
                    This execution is run because the results of saved search "
                    <a href="">
                        <i>{execution.action.savedSearch.description}</i>
                    </a>
                    " changed.
                </div>
            )}
            <h2>Action definition</h2>
            <MonacoSettingsEditor
                isLightTheme={isLightTheme}
                readOnly={true}
                language="json"
                value={actionDefinition}
                height={200}
                className="mb-3"
            ></MonacoSettingsEditor>
            <h2>Action status</h2>
            <p>
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
            </p>
            <div className="progress">
                <div className="progress-bar" style={{ width: (449 / 1500) * 100 + '%' }}>
                    &nbsp;
                </div>
            </div>
            <p className="text-center w-100">449 / 1500 jobs ran</p>
            <h2>Action jobs</h2>
            <ul className="list-group mb-3">
                {execution.jobs.nodes.map(actionJob => (
                    <ActionJob isLightTheme={isLightTheme} actionJob={actionJob} key={actionJob.id} />
                ))}
            </ul>
            <h2>Action summary</h2>
            <p>
                The action execution finished, some tasks failed. <span className="badge badge-secondary">15</span>{' '}
                patches have been created.
            </p>
            <div className="mb-3">
                <button type="button" className="btn btn-primary">
                    Create campaign
                </button>
                <button type="button" className="btn btn-primary ml-2">
                    Update existing campaign
                </button>
            </div>
        </>
    )
}
