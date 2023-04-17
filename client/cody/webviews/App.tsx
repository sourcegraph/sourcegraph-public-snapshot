import { useCallback, useEffect, useState } from 'react'

import './App.css'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { ChatHistory, ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Configuration } from '@sourcegraph/cody-shared/src/configuration'

import { Chat } from './Chat'
import { Debug } from './Debug'
import { Header } from './Header'
import { LoadingPage } from './LoadingPage'
import { Login } from './Login'
import { NavBar, View } from './NavBar'
import { Recipes } from './Recipes'
import { Settings } from './Settings'
import { UserHistory } from './UserHistory'
import { vscodeAPI } from './utils/VSCodeApi'

export function App(): React.ReactElement {
    const [config, setConfig] = useState<Pick<Configuration, 'debug' | 'serverEndpoint'> | null>(null)
    const [debugLog, setDebugLog] = useState(['No data yet'])
    const [view, setView] = useState<View | undefined>()
    const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
    const [transcript, setTranscript] = useState<ChatMessage[]>([])
    const [isValidLogin, setIsValidLogin] = useState<boolean>()
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [userHistory, setUserHistory] = useState<ChatHistory | null>(null)
    const [contextStatus, setContextStatus] = useState<ChatContextStatus | null>(null)

    useEffect(() => {
        vscodeAPI.onMessage(message => {
            switch (message.type) {
                case 'transcript': {
                    if (message.isMessageInProgress) {
                        const msgLength = message.messages.length - 1
                        setTranscript(message.messages.slice(0, msgLength))
                        setMessageInProgress(message.messages[msgLength])
                    } else {
                        setTranscript(message.messages)
                        setMessageInProgress(null)
                    }
                    break
                }
                case 'config':
                    setConfig(message.config)
                    setView(message.config.hasAccessToken ? 'chat' : 'login')
                    break
                case 'login':
                    setIsValidLogin(message.isValid)
                    setView(message.isValid ? 'chat' : 'login')
                    break
                case 'showTab':
                    if (message.tab === 'chat') {
                        setView('chat')
                    }
                    break
                case 'debug':
                    setDebugLog([...debugLog, message.message])
                    break
                case 'history':
                    setInputHistory(message.messages?.input ?? [])
                    setUserHistory(message.messages?.chat ?? null)
                    break
                case 'contextStatus':
                    setContextStatus(message.contextStatus)
                    break
                case 'view':
                    setView(message.messages)
                    break
            }
        })

        vscodeAPI.postMessage({ command: 'initialized' })
        // The dependencies array is empty to execute the callback only on component mount.
    }, [debugLog])

    const onLogin = useCallback((token: string, endpoint: string) => {
        if (!token || !endpoint) {
            return
        }
        setIsValidLogin(undefined)
        vscodeAPI.postMessage({ command: 'settings', serverEndpoint: endpoint, accessToken: token })
    }, [])

    const onLogout = useCallback(() => {
        vscodeAPI.postMessage({ command: 'removeToken' })
        setView('login')
    }, [setView])

    const onResetClick = useCallback(() => {
        setView('chat')
        setDebugLog([])
        setFormInput('')
        setMessageInProgress(null)
        setTranscript([])
        vscodeAPI.postMessage({ command: 'reset' })
    }, [setView, setMessageInProgress, setTranscript, setDebugLog])

    if (!view) {
        return <LoadingPage />
    }

    return (
        <div className="outer-container">
            <Header />
            {view === 'login' && (
                <Login onLogin={onLogin} isValidLogin={isValidLogin} serverEndpoint={config?.serverEndpoint} />
            )}
            {view && view !== 'login' && (
                <NavBar
                    view={view}
                    setView={setView}
                    devMode={Boolean(config?.debug)}
                    onResetClick={onResetClick}
                    showResetButton={transcript.length > 0}
                />
            )}
            {view === 'debug' && config?.debug && <Debug debugLog={debugLog} />}
            {view === 'history' && (
                <UserHistory
                    userHistory={userHistory}
                    setUserHistory={setUserHistory}
                    setInputHistory={setInputHistory}
                />
            )}
            {view === 'recipes' && <Recipes />}
            {view === 'settings' && (
                <Settings setView={setView} onLogout={onLogout} serverEndpoint={config?.serverEndpoint} />
            )}
            {view === 'chat' && (
                <Chat
                    messageInProgress={messageInProgress}
                    transcript={transcript}
                    contextStatus={contextStatus}
                    formInput={formInput}
                    setFormInput={setFormInput}
                    inputHistory={inputHistory}
                    setInputHistory={setInputHistory}
                />
            )}
        </div>
    )
}
