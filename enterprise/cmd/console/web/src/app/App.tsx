import './App.css'

import { useObservable } from '@sourcegraph/wildcard'
import React, { useMemo } from 'react'
import { Link, Redirect, Route, Switch } from 'react-router-dom'

import { InstanceList } from '../instances/instanceList/InstanceList'
import { SignupPage } from '../trialStartFlow/steps/1-signup/SignupPage'
import { newAPIClient } from '../model/apiClient'
import { NewInstancePage } from '../trialStartFlow/steps/2-instance/NewInstancePage'
import { WaitForInstancePage } from '../trialStartFlow/steps/3-wait/WaitForInstancePage'
import { ConsolePage } from '../console/ConsolePage'

export const App: React.FunctionComponent = () => {
    const apiClient = useMemo(() => newAPIClient(), [])
    const data = useObservable(useMemo(() => apiClient.getData(), [apiClient]))

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
                    <Route path="/instances/wait">
                        <WaitForInstancePage />
                    </Route>
                    <Route path="/">
                        <ConsolePage data={data} />
                    </Route>
                </Switch>
            )}
        </>
    )
}
