import './App.css'

import { useObservableState } from 'observable-hooks'
import React, { useMemo } from 'react'
import { Link, Navigate, Route, Routes, useLocation } from 'react-router-dom'

import { Instances } from '../instances/Instances'
import { NewInstance } from '../instances/NewInstance'
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
                <Routes>
                    <Route path="/new-instance" element={<NewInstance />} />
                    <Route
                        path="/sign-in"
                        element={
                            <p>
                                Sign in or <Link to="/new-instance">create new instance</Link>
                            </p>
                        }
                    />
                    <Route path="*" element={<Navigate to="/sign-in" />} />
                </Routes>
            ) : (
                <Routes>
                    <Route path="/" element={<p>hello8</p>} />
                    <Route path="/instances" element={<Instances instances={data.instances} className="content" />} />
                </Routes>
            )}
        </>
    )
}
