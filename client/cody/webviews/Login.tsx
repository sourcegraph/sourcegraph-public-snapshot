import { useState } from 'react'

import { TextFieldType } from '@vscode/webview-ui-toolkit/dist/text-field'
import { VSCodeTextField, VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { Terms } from './About'

import './Login.css'

interface LoginProps {
    isValidLogin?: boolean
    onLogin: (token: string, endpoint: string) => void
}

export const Login: React.FunctionComponent<React.PropsWithChildren<LoginProps>> = ({ isValidLogin, onLogin }) => {
    const [termsAccepted, setTermsAccepted] = useState(false)
    const [token, setToken] = useState<string>('')
    const [endpoint, setEndpoint] = useState('')

    return (
        <div className="inner-container">
            <div className="non-transcript-container">
                {termsAccepted ? (
                    <div className="container-getting-started">
                        <p>Access Token</p>
                        <VSCodeTextField
                            value={token}
                            placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                            className="w-100"
                            type={TextFieldType.password}
                            onInput={(e: any) => setToken(e.target.value)}
                        />
                        <p>Sourcegraph Instance</p>
                        <VSCodeTextField
                            value={endpoint}
                            placeholder="ex https://example.sourcegraph.com"
                            className="w-100"
                            autofocus={true}
                            onInput={(e: any) => setEndpoint(e.target.value)}
                        />
                        <VSCodeButton className="login-button" type="button" onClick={() => onLogin(token, endpoint)}>
                            Login
                        </VSCodeButton>
                        {isValidLogin === false && (
                            <p className="invalid-login-message">
                                Invalid login credentials. Please check that you have entered the correct instance URL
                                and a valid access token.
                            </p>
                        )}
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
