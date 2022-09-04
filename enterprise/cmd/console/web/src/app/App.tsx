import './App.css'

import { useObservableState } from 'observable-hooks'
import React, { useMemo } from 'react'
import { Route, Routes, useLocation } from 'react-router-dom'

import { Instances } from '../instances/Instances'
import { newAPIClient } from '../model/apiClient'
import { Header } from './Header'

export const App: React.FunctionComponent = () => {
    const apiClient = useMemo(() => newAPIClient(), [])
    const data = useObservableState(useMemo(() => apiClient.getData(), [apiClient]))

    const location = useLocation()
    return (
        <>
            <Header data={data} />
            {data === undefined ? (
                <p>Loading...</p>
            ) : data.user === null ? (
                <p>Sign in</p>
            ) : (
                <Routes>
                    <Route path="/" element={<p>hello8</p>} />
                    <Route path="/instances" element={<Instances instances={data.instances} className="content" />} />
                </Routes>
            )}
        </>
    )
}
