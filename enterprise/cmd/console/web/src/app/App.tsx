import './App.css'

import React, { useMemo } from 'react'

import { googleAuthProvider } from '../services/google/authProvider'
import { googleTasksTaskProvider } from '../services/google/googleTasks'
import { TaskProvider } from '../tasks/taskProvider'
import { Tasks } from '../tasks/Tasks'
import { AuthProvider } from './auth'
import { Header } from './Header'
import { useSettings } from './useSettings'

export const App: React.FunctionComponent = () => {
    const authProviders: AuthProvider[] = useMemo(() => [googleAuthProvider], [])
    const taskProviders: TaskProvider[] = useMemo(() => [googleTasksTaskProvider], [])
    const [settings, setSettings] = useSettings()
    return (
        <>
            <Header settings={settings} setSettings={setSettings} authProviders={authProviders} />
            <Tasks taskProviders={taskProviders} settings={settings} className="content" />
        </>
    )
}
