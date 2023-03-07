import { useCallback, useEffect, useState } from 'react'

import './App.css'
import { Header, About, LoadingPage } from './About'
import { Chat } from './Chat'
import { vscodeAPI, MessageFromWebview } from './utils/vscodeAPI'
import Login from './Login'
import Recipes from './Recipes'
import { NavBar } from './NavBar'
import { ChatMessage } from '../commands/ChatViewProvider'

function App() {
    const devMode = false
    const [view, setView] = useState(10)
    const [token, setToken] = useState<string | null>(null)
    const [isLoggedIn, setIsLoggedIn] = useState(devMode ? true : false)
    const [endpoint, setEndpoint] = useState('https://cody.sgdev.org')
    const [userInput, setUserInput] = useState('')
    const [transcript, setTranscript] = useState<ChatMessage | null>(null)
    const [transcripts, setTranscripts] = useState<ChatMessage[]>([])
    const [inConversation, setInConversation] = useState(true)
    const [codyIsTyping, setCodyIsTyping] = useState(false)

    useEffect(() => {
        // at startup
        if (token === null && view === 10) {
            vscodeAPI.postMessage({
                command: 'get-token',
                value: 'on-startup',
            } as MessageFromWebview)
            setToken('')
            // listen to messages from extension client
            vscodeAPI.onMessage(message => {
                // get transcripts from extension client
                if (message.data && message.data.type === 'transcripts') {
                    setTranscripts(message.data.messages)
                    setCodyIsTyping(message.data.messageInProgress)
                }
                // get token from extension client
                if (message.data && message.data.type === 'auth' && message.data.value) {
                    setToken(message.data.value)
                    setIsLoggedIn(true)
                    setView(1)
                }
            })
        }
    }, [token, userInput, endpoint, view, inConversation])

    const onTokenSubmit = useCallback(() => {
        if (!token || !endpoint) {
            return
        }
        // Create wsclient
        vscodeAPI.postMessage({
            command: 'settings',
            serverURL: endpoint,
            accessToken: token,
        } as MessageFromWebview)
        // set token
        vscodeAPI.postMessage({
            command: 'set-token',
            value: token,
        } as MessageFromWebview)
        // Set login view
        setIsLoggedIn(true)
        setView(1)
    }, [token, endpoint])

    const onLogoutClick = useCallback(() => {
        setToken('')
        vscodeAPI.postMessage({
            command: 'remove-token',
            value: 'remove-token',
        } as MessageFromWebview)
        setToken('')
        setInConversation(false)
        setIsLoggedIn(false)
        setView(10)
    }, [])

    const onResetClick = useCallback(() => {
        setView(1)
        setUserInput('')
        setTranscript(null)
        setTranscripts([])
        setInConversation(false)
        vscodeAPI.postMessage({
            command: 'reset',
        } as MessageFromWebview)
    }, [])

    if (token === null) {
        return <LoadingPage />
    }

    return (
        <div className="outer-container">
            <Header showResetBtn={isLoggedIn} onResetClick={onResetClick} />
            {isLoggedIn && view !== 10 && <NavBar onLogoutClick={onLogoutClick} view={view} setView={setView} />}
            {!isLoggedIn && <Login setToken={setToken} setEndpoint={setEndpoint} onTokenSubmit={onTokenSubmit} />}
            {isLoggedIn && view === 2 && <Recipes />}
            {isLoggedIn && view === 3 && <About />}
            {isLoggedIn && view === 1 && (
                <Chat
                    transcript={transcript}
                    transcripts={transcripts}
                    setTranscript={setTranscript}
                    setTranscripts={setTranscripts}
                    setInConversation={setInConversation}
                    setUserInput={setUserInput}
                    codyIsTyping={codyIsTyping}
                    setCodyIsTyping={setCodyIsTyping}
                />
            )}
        </div>
    )
}

export default App
