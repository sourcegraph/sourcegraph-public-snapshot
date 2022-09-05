import './App.css'

import { useObservable } from '@sourcegraph/wildcard'
import React, { useMemo } from 'react'
import { Link, Redirect, Route, Switch } from 'react-router-dom'

import { Instances } from '../instances/Instances'
import { SignupPage } from '../signup/SignupPage'
import { newAPIClient } from '../model/apiClient'

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
                    <Route path="/">
                        <p>hello8</p>
                    </Route>
                    <Route path="/instances">
                        <Instances instances={data.instances} className="content" />
                    </Route>
                </Switch>
            )}
        </>
    )
}
