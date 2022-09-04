import './App.css'

import { useObservableState } from 'observable-hooks'
import React, { useMemo } from 'react'

import { Instances } from '../instances/Instances'
import { newAPIClient } from '../model/apiClient'
import { Header } from './Header'

export const App: React.FunctionComponent = () => {
    const apiClient = useMemo(() => newAPIClient(), [])
    const data = useObservableState(useMemo(() => apiClient.getData(), [apiClient]))
    return (
        <>
            <Header data={data} />
            {data === undefined ? (
                <p>Loading...</p>
            ) : data.user === null ? (
                <p>Sign in</p>
            ) : (
                <Instances instances={data.instances} className="content" />
            )}
        </>
    )
}
