import './App.css'

import { useObservable } from '@sourcegraph/wildcard'
import React, { useMemo } from 'react'
import { Link, Redirect, Route, Switch } from 'react-router-dom'

import { Instances } from '../instances/Instances'
import { NewInstance } from '../instances/NewInstance'
import { newAPIClient } from '../model/apiClient'
import { Header } from './Header'

export const App: React.FunctionComponent = () => {
    const apiClient = useMemo(() => newAPIClient(), [])
    const data = useObservable(useMemo(() => apiClient.getData(), [apiClient]))

    return (
        <>
            <Header data={data} />
            {data === undefined ? (
                <p>Loading...</p>
            ) : data.user === null ? (
                <Switch>
                    <Route path="/new-instance">
                        <NewInstance />
                    </Route>
                    <Route path="/sign-in">
                        <p>
                            Sign in or <Link to="/new-instance">create new instance</Link>
                        </p>
                    </Route>
                    <Route path="*">
                        <Redirect to="/sign-in" />
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
