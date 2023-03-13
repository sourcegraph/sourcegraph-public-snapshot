import { useState } from 'react'

import { VSCodeTextField, VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { Terms } from './About'

import './Login.css'

interface LoginProps {
	onLogin: (token: string, endpoint: string) => void
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({ onLogin }) => {
	const [termsAccepted, setTermsAccepted] = useState(false)
	const [token, setToken] = useState<string>('')
	const [endpoint, setEndpoint] = useState('https://cody.sgdev.org')

	return (
		<div className="inner-container">
			<div className="non-transcript-container">
				{termsAccepted ? (
					<div className="container-getting-started">
						<h1>Login</h1>
						<p>Access Token</p>
						<VSCodeTextField
							value={token}
							placeholder="12345a678b91c2d34e567"
							className="w-100"
							type="password"
							onInput={(e: any) => setToken(e.target.value)}
						/>
						<p>Endpoint for Cody</p>
						<VSCodeTextField
							value={endpoint}
							placeholder="https://cody.sgdev.org"
							className="w-100"
							autofocus={true}
							onInput={(e: any) => setEndpoint(e.target.value)}
						/>
						<VSCodeButton className="login-button" type="button" onClick={() => onLogin(token, endpoint)}>
							Login
						</VSCodeButton>
					</div>
				) : (
					<div className="container-getting-started">
						<Terms setTermsAccepted={setTermsAccepted} />
					</div>
				)}
			</div>
		</div>
	)
}
