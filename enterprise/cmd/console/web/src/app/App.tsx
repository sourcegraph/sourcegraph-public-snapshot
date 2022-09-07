import './App.css'

import { useObservable } from '@sourcegraph/wildcard'
import React, { useMemo } from 'react'
import { Link, Redirect, Route, Switch, useLocation } from 'react-router-dom'

import { InstanceList } from '../instances/instanceListPage/InstanceList'
import { SignupPage } from '../trialStartFlow/steps/1-signup/SignupPage'
import { newAPIClient } from '../model/apiClient'
import { NewInstancePage } from '../trialStartFlow/steps/2-instance/NewInstancePage'
import { WaitForInstancePage } from '../trialStartFlow/steps/3-wait/WaitForInstancePage'
import { ConsolePage } from '../console/ConsolePage'
import { InstanceListPage } from '../instances/instanceListPage/InstanceListPage'
import { ConsoleLayout } from '../console/ConsoleLayout'
import { InstanceManagePage } from '../instances/instanceManagePage/InstanceManagePage'

export const App: React.FunctionComponent = () => {
    const apiClient = useMemo(() => newAPIClient(), [])
    const data = useObservable(useMemo(() => apiClient.getData(), [apiClient]))

    const location = useLocation()
    console.log('location', location.pathname)

    return (
        <>
            {data === undefined ? null : data.user === null ? (
                <Switch>
                    <Route path="/signup">
                        <SignupPage />
                    </Route>
                    <Route path="*">
                        <Redirect to="/signup" />
                    </Route>
                </Switch>
            ) : (
                <Switch>
                    <Route path="/new-instance">
                        <NewInstancePage />
                    </Route>
                    <Route path="/instances/:id/wait">
                        <WaitForInstancePage />
                    </Route>
                    <Route
                        path="/instances"
                        render={({ match: { path } }) => (
                            <ConsoleLayout data={data}>
                                <Switch>
                                    <Route
                                        path={`${path}/:id`}
                                        render={({
                                            match: {
                                                params: { id },
                                            },
                                        }) => {
                                            const instance = data.instances.find(instance => instance.id === id)
                                            return instance ? (
                                                <InstanceManagePage instance={instance} />
                                            ) : (
                                                <p>404 Not Found</p>
                                            )
                                        }}
                                    />
                                    <Route path={path} exact={true}>
                                        <InstanceListPage data={data} />
                                    </Route>
                                </Switch>
                            </ConsoleLayout>
                        )}
                    />
                    <Route path="/" exact={true}>
                        <Redirect to="/instances" />
                    </Route>
                </Switch>
            )}
        </>
    )
}
