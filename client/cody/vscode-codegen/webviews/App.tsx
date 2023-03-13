import { useCallback, useEffect, useState } from 'react'

import './App.css'

import { About } from './About'
import { Chat } from './Chat'
import { Header } from './Header'
import { LoadingPage } from './LoadingPage'
import { Login } from './Login'
import { NavBar } from './NavBar'
import { Recipes } from './Recipes'
import { Settings } from './Settings'
import { ChatMessage, View } from './utils/types'
import { vscodeAPI, WebviewMessage } from './utils/vscodeAPI'

function App(): React.ReactElement {
	const [devMode, setDevMode] = useState(false)
	const [view, setView] = useState<View | undefined>()
	const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
	const [transcript, setTranscript] = useState<ChatMessage[]>([])

	useEffect(() => {
		vscodeAPI.onMessage(message => {
			// Get chat transcript from extension.
			if (message.data && message.data.type === 'transcript') {
				setTranscript(message.data.messages)
				setMessageInProgress(message.data.messageInProgress)
			}
			// Get the token from the extension.
			if (message.data && message.data.type === 'token') {
				const hasToken = !!message.data.value
				setView(hasToken ? 'chat' : 'login')
				setDevMode(message.data.mode === 'development')
			}
		})

		vscodeAPI.postMessage({ command: 'initialized' } as WebviewMessage)
		// The dependencies array is empty to execute the callback only on component mount.
	}, [])

	const onLogin = useCallback(
		(token: string, endpoint: string) => {
			if (!token || !endpoint) {
				return
			}
			// Create wsclient.
			vscodeAPI.postMessage({ command: 'settings', serverURL: endpoint, accessToken: token } as WebviewMessage)
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
		setMessageInProgress(null)
		setTranscript([])
		vscodeAPI.postMessage({ command: 'reset' } as WebviewMessage)
	}, [setView, setMessageInProgress, setTranscript])

	if (!view) {
		return <LoadingPage />
	}

	return (
		<div className="outer-container">
			<Header showResetButton={view && view !== 'login'} onResetClick={onResetClick} />
			{view === 'login' && <Login onLogin={onLogin} />}
			{view && view !== 'login' && <NavBar view={view} setView={setView} />}
			{view === 'recipes' && <Recipes />}
			{view === 'about' && <About />}
			{view === 'settings' && <Settings onLogout={onLogout} />}
			{view === 'chat' && (
				<Chat
					messageInProgress={messageInProgress}
					transcript={transcript}
					setMessageInProgress={setMessageInProgress}
					setTranscript={setTranscript}
					devMode={devMode}
				/>
			)}
		</div>
	)
}

export default App
