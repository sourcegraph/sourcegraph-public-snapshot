import './Header.css'

import React from 'react'

import { AuthProvider } from './auth'
import { SettingsProps } from './useSettings'

export const Header: React.FunctionComponent<
    {
        authProviders: AuthProvider[]
    } & SettingsProps
> = ({ settings, setSettings, authProviders }) => (
    <header id="header">
        <div className="container">
            <nav className="main">
                <h1 id="logo">
                    <a href="/">Sourcegraph Console</a>
                </h1>
            </nav>
            <div style={{ flex: 1 }} />
            <nav className="user">
                <ul className="accounts">
                    {authProviders.map(({ name, signInComponent: SignInComponent }) => (
                        <li key={name}>
                            <SignInComponent settings={settings} setSettings={setSettings} />
                        </li>
                    ))}
                </ul>
            </nav>
        </div>
    </header>
)
