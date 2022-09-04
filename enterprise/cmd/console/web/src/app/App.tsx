import './App.css'

import { useObservable, useObservableGetState, useObservableState } from 'observable-hooks'
import React, { useMemo } from 'react'

import { Instances } from '../instances/Instances'
import { newAPIClient } from '../model/apiClient'
import { googleAuthProvider } from '../services/google/authProvider'
import { AuthProvider } from './auth'
import { Header } from './Header'
import { useSettings } from './useSettings'

export const App: React.FunctionComponent = () => {
    const authProviders: AuthProvider[] = useMemo(() => [googleAuthProvider], [])

    const apiClient = useMemo(() => newAPIClient(), [])
    const data = useObservableState(useMemo(() => apiClient.getData(), [apiClient]))

    const [settings, setSettings] = useSettings()
    return (
        <>
            <Header settings={settings} setSettings={setSettings} authProviders={authProviders} />
            <Instances instances={data?.instances} className="content" />
        </>
    )
}
