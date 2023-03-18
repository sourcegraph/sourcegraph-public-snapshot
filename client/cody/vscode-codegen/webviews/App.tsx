import { useCallback, useEffect, useState } from 'react'

import './App.css'

import { UserLocalHistory } from '../src/editor/LocalStorageProvider'

import { About } from './About'
import { Chat } from './Chat'
import { Debug } from './Debug'
import { Header } from './Header'
import { LoadingPage } from './LoadingPage'
import { Login } from './Login'
import { NavBar } from './NavBar'
import { Recipes } from './Recipes'
import { Settings } from './Settings'
import { ChatMessage, View } from './utils/types'
import { vscodeAPI, WebviewMessage } from './utils/VSCodeApi'

function App(): React.ReactElement {
    const [devMode, setDevMode] = useState(false)
    const [debugLog, setDebugLog] = useState(['Debug mode on.'])
    const [view, setView] = useState<View | undefined>()
    const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
    const [transcript, setTranscript] = useState<ChatMessage[]>([])
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [userHistory, setUserHistory] = useState<UserLocalHistory | null>(null)

    useEffect(() => {
        vscodeAPI.onMessage(message => {
            switch (message.data.type) {
                // Get chat transcript from extension.
                case 'transcript':
                    setTranscript(message.data.messages)
                    setMessageInProgress(message.data.messageInProgress)
                    break
                // Get the token from the extension.
                case 'token':
                    const hasToken = !!message.data.value
                    setView(hasToken ? 'chat' : 'login')
                    setDevMode(message.data.mode === 'development')
                    break
                case 'debug':
                    setDebugLog([...debugLog, message.data.message])
                    break
                case 'history':
                    console.log(message.data.messages)
                    setInputHistory(message.data.messages.input)
                    setUserHistory(message.data.messages.chat)
                    break
            }
        })
        vscodeAPI.postMessage({ command: 'initialized' } as WebviewMessage)
    }, []) // The dependencies array is empty to execute the callback only on component mount.

    const onLogin = useCallback(
        (token: string, endpoint: string) => {
            if (!token || !endpoint) {
                return
            }
            vscodeAPI.postMessage({
                command: 'settings',
                serverEndpoint: endpoint,
                accessToken: token,
            } as WebviewMessage)
            // Set token.
            vscodeAPI.postMessage({ command: 'setToken', value: token } as WebviewMessage)
            setView('chat')
        },
        [setView]
    )

    const onLogout = useCallback(() => {
        vscodeAPI.postMessage({ command: 'removeToken' } as WebviewMessage)
        setView('login')
    }, [setView])

    const onResetClick = useCallback(() => {
        setView('chat')
        setDebugLog([])
        setFormInput('')
        setMessageInProgress(null)
        setTranscript([])
        vscodeAPI.postMessage({ command: 'reset' } as WebviewMessage)
    }, [setView, setMessageInProgress, setTranscript, setDebugLog])

    if (!view) {
        return <LoadingPage />
    }

    return (
        <div className="outer-container">
            <Header showResetButton={view && view !== 'login'} onResetClick={onResetClick} />
            {view === 'login' && <Login onLogin={onLogin} />}
            {view && view !== 'login' && <NavBar view={view} setView={setView} devMode={devMode} />}
            {view === 'about' && <About />}
            {view === 'debug' && devMode && <Debug debugLog={debugLog} />}
            {view === 'recipes' && <Recipes />}
            {view === 'settings' && <Settings setView={setView} onLogout={onLogout} userHistory={userHistory} />}
            {view === 'chat' && (
                <Chat
                    messageInProgress={messageInProgress}
                    transcript={transcript}
                    setMessageInProgress={setMessageInProgress}
                    setTranscript={setTranscript}
                    formInput={formInput}
                    setFormInput={setFormInput}
                    inputHistory={inputHistory}
                    setInputHistory={setInputHistory}
                    onResetClick={onResetClick}
                />
            )}
        </div>
    )
}

export default App
