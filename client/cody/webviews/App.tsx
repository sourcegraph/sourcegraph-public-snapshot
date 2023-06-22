import { useCallback, useEffect, useState } from 'react'

import './App.css'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { ChatHistory, ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Configuration } from '@sourcegraph/cody-shared/src/configuration'

import { AuthStatus, LocalEnv, isLoggedIn } from '../src/chat/protocol'

import { Chat } from './Chat'
import { Debug } from './Debug'
import { Header } from './Header'
import { LoadingPage } from './LoadingPage'
import { Login } from './Login'
import { NavBar, View } from './NavBar'
import { Recipes } from './Recipes'
import { Settings } from './Settings'
import { UserHistory } from './UserHistory'
import type { VSCodeWrapper } from './utils/VSCodeApi'

export const App: React.FunctionComponent<{ vscodeAPI: VSCodeWrapper }> = ({ vscodeAPI }) => {
    const [config, setConfig] = useState<(Pick<Configuration, 'debugEnable' | 'serverEndpoint'> & LocalEnv) | null>(
        null
    )
    const [debugLog, setDebugLog] = useState<string[]>([])
    const [view, setView] = useState<View | undefined>()
    const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
    const [messageBeingEdited, setMessageBeingEdited] = useState<boolean>(false)
    const [transcript, setTranscript] = useState<ChatMessage[]>([])
    const [authStatus, setAuthStatus] = useState<AuthStatus>()
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [userHistory, setUserHistory] = useState<ChatHistory | null>(null)
    const [contextStatus, setContextStatus] = useState<ChatContextStatus | null>(null)
    const [errorMessages, setErrorMessages] = useState<string[]>([])
    const [suggestions, setSuggestions] = useState<string[] | undefined>()
    const [isAppInstalled, setIsAppInstalled] = useState<boolean>(false)

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
                    setAuthStatus(message.authStatus)
                    setView(isLoggedIn(message.authStatus) ? 'chat' : 'login')
                    break
                case 'login':
                    setAuthStatus(message.authStatus)
                    setIsAuthenticated(isLoggedIn(message.authStatus))
                    setView(isLoggedIn(message.authStatus) ? 'chat' : 'login')
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
                case 'errors':
                    setErrorMessages([...errorMessages, message.errors].slice(-5))
                    setDebugLog([...debugLog, message.errors])
                    break
                case 'view':
                    setView(message.messages)
                    break
                case 'suggestions':
                    setSuggestions(message.suggestions)
                    break
                case 'app-state':
                    setIsAppInstalled(message.isInstalled)
                    break
            }
        })

        vscodeAPI.postMessage({ command: 'initialized' })
        // The dependencies array is empty to execute the callback only on component mount.
    }, [debugLog, errorMessages, vscodeAPI])

    const onLogout = useCallback(() => {
        setIsAuthenticated(false)
        vscodeAPI.postMessage({ command: 'auth', type: 'signout' })
    }, [vscodeAPI])

    if (!view) {
        return <LoadingPage />
    }

    return (
        <div className="outer-container">
            <Header />
            {view === 'login' && !isAuthenticated ? (
                <Login
                    authStatus={authStatus}
                    serverEndpoint={config?.serverEndpoint}
                    isAppInstalled={isAppInstalled}
                    isAppRunning={config?.isAppRunning}
                    vscodeAPI={vscodeAPI}
                    appOS={config?.os}
                    appArch={config?.arch}
                    callbackScheme={config?.uriScheme}
                    isAppConnectEnabled={config?.isAppConnectEnabled}
                />
            ) : (
                <>
                    <NavBar view={view} setView={setView} devMode={Boolean(config?.debugEnable)} />
                    {errorMessages && <ErrorBanner errors={errorMessages} setErrors={setErrorMessages} />}
                    {view === 'debug' && config?.debugEnable && <Debug debugLog={debugLog} />}
                    {view === 'history' && (
                        <UserHistory
                            userHistory={userHistory}
                            setUserHistory={setUserHistory}
                            setInputHistory={setInputHistory}
                            setView={setView}
                            vscodeAPI={vscodeAPI}
                        />
                    )}
                    {view === 'recipes' && <Recipes vscodeAPI={vscodeAPI} />}
                    {view === 'settings' && (
                        <Settings
                            onLogout={onLogout}
                            endpoint={authStatus?.endpoint || config?.serverEndpoint}
                            version={config?.extensionVersion}
                        />
                    )}
                    {view === 'chat' && (
                        <Chat
                            messageInProgress={messageInProgress}
                            messageBeingEdited={messageBeingEdited}
                            setMessageBeingEdited={setMessageBeingEdited}
                            transcript={transcript}
                            contextStatus={contextStatus}
                            formInput={formInput}
                            setFormInput={setFormInput}
                            inputHistory={inputHistory}
                            setInputHistory={setInputHistory}
                            vscodeAPI={vscodeAPI}
                            suggestions={suggestions}
                            setSuggestions={setSuggestions}
                        />
                    )}
                </>
            )}
        </div>
    )
}

const ErrorBanner: React.FunctionComponent<{ errors: string[]; setErrors: (errors: string[]) => void }> = ({
    errors,
    setErrors,
}) => (
    <div className="error-container">
        {errors.map((error, i) => (
            <div key={i} className="error">
                <span>{error}</span>
                <button type="button" className="close-btn" onClick={() => setErrors(errors.filter(e => e !== error))}>
                    ×
                </button>
            </div>
        ))}
    </div>
)
