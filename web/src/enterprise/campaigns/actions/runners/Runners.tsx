import React, { useMemo } from 'react'
import { PageTitle } from '../../../../components/PageTitle'
import { Collapsible } from '../../../../components/Collapsible'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import RefreshIcon from 'mdi-react/RefreshIcon'
import { useObservable } from '../../../../../../shared/src/util/useObservable'
import { queryActionJobs } from './backend'
import { ActionJobState } from '../../../../../../shared/src/graphql/schema'

interface Props {}

export const Runners: React.FunctionComponent<Props> = () => {
    const jobsConnection = useObservable(
        // don't need the action jobs, just want the totalCount
        useMemo(() => queryActionJobs({ state: ActionJobState.PENDING, first: 1 }), [])
    )
    return (
        <>
            <PageTitle title="Manage runners" />
            <h1>Manage runners</h1>
            <p>Register a new runner with Sourcegraph</p>
            <div className="alert alert-info mb-4">
                <p>Use the below token to register your runner to this Sourcegraph instance</p>
                <input className="form-control mb-2" readOnly={true} value="KK3DK99AA1291S8" />
                <div>
                    <code>
                        src actions runner --sourcegraph-url={window.location.protocol}//{window.location.host}{' '}
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
            <h2>Active runners</h2>
            <ul className="list-group mb-4">
                <li className="list-group-item">
                    <Collapsible
                        title={
                            <div className="ml-2 d-flex justify-content-between align-content-center">
                                <div>
                                    <h3>Runner sg-3919</h3>
                                    <p className="text-monospace mb-0">
                                        <small>macOS 10.15.2, Docker 19.06.03, 32 CPU, 64 GB RAM</small>
                                    </p>
                                </div>
                                <CheckboxBlankCircleIcon
                                    data-tooltip="Runner is online"
                                    className="d-block text-success"
                                />
                            </div>
                        }
                        titleClassName="flex-grow-1"
                        wholeTitleClickable={false}
                    >
                        <h5>Running jobs</h5>
                        <ul className="list-group">
                            <li className="list-group-item">
                                <a href="#">Job #19239</a>
                                <RefreshIcon className="icon-loader" />
                            </li>
                        </ul>
                    </Collapsible>
                </li>
            </ul>
            <h2>Active jobs</h2>
            {/* TODO: Endpoint to receive jobs by status */}
            <p>Current backlog: {jobsConnection?.totalCount} jobs pending</p>
            <ul className="list-group">
                <li className="list-group-item">Job 1</li>
            </ul>
        </>
    )
}
