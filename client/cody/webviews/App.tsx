import { useCallback, useEffect, useState } from 'react'

import './App.css'

import { About } from './About'
import { Chat } from './Chat'
import { Debug } from './Debug'
import { Header } from './Header'
import { LoadingPage } from './LoadingPage'
import { Login } from './Login'
import { NavBar } from './NavBar'
import { Recipes } from './Recipes'
import { Settings } from './Settings'
import { UserHistory } from './UserHistory'
import { ChatHistory, ChatMessage, View } from './utils/types'
import { vscodeAPI } from './utils/VSCodeApi'

function App(): React.ReactElement {
    const [devMode, setDevMode] = useState(false)
    const [debugLog, setDebugLog] = useState(['No data yet'])
    const [view, setView] = useState<View | undefined>()
    const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
    const [transcript, setTranscript] = useState<ChatMessage[]>([])
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [userHistory, setUserHistory] = useState<ChatHistory | null>(null)

    useEffect(() => {
        vscodeAPI.onMessage(message => {
            switch (message.data.type) {
                case 'transcript': {
                    if (message.data.isMessageInProgress) {
                        const msgLength = message.data.messages.length - 1
                        setTranscript(message.data.messages.slice(0, msgLength))
                        setMessageInProgress(message.data.messages[msgLength])
                    } else {
                        setTranscript(message.data.messages)
                        setMessageInProgress(null)
                    }
                    break
                }
                case 'token':
                    // Get the token from the extension.
                    const hasToken = !!message.data.value
                    setView(hasToken ? 'chat' : 'login')
                    setDevMode(message.data.mode === 'development')
                    break
                case 'showTab':
                    if (message.data.tab === 'chat') {
                        setView('chat')
                    }
                    break
                case 'debug':
                    setDebugLog([...debugLog, message.data.message])
                    break
                case 'history':
                    setInputHistory(message.data.messages.input)
                    setUserHistory(message.data.messages.chat)
                    break
            }
        })
        vscodeAPI.postMessage({ command: 'initialized' })
        // The dependencies array is empty to execute the callback only on component mount.
    }, [])

    const onLogin = useCallback(
        (token: string, endpoint: string) => {
            if (!token || !endpoint) {
                return
            }
            vscodeAPI.postMessage({ command: 'settings', serverEndpoint: endpoint, accessToken: token })
            setView('chat')
        },
        [setView]
    )

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
            <Header showResetButton={view && view !== 'login'} onResetClick={onResetClick} />
            {view === 'login' && <Login onLogin={onLogin} />}
            {view && view !== 'login' && <NavBar view={view} setView={setView} devMode={devMode} />}
            {view === 'about' && <About />}
            {view === 'debug' && devMode && <Debug debugLog={debugLog} />}
            {view === 'recipes' && <Recipes />}
            {view === 'settings' && <Settings setView={setView} onLogout={onLogout} />}
            {view === 'history' && <UserHistory userHistory={userHistory} />}
            {view === 'chat' && (
                <Chat
                    messageInProgress={messageInProgress}
                    transcript={transcript}
                    formInput={formInput}
                    setFormInput={setFormInput}
                    inputHistory={inputHistory}
                    setInputHistory={setInputHistory}
                />
            )}
        </div>
    )
}

export default App
