import './Header.css'

import React from 'react'

import { ConsoleData } from '../model'

export const Header: React.FunctionComponent<{
    data: ConsoleData | undefined
}> = ({ data }) => (
    <header id="header">
        <div className="container">
            <nav className="main">
                <h1 id="logo">
                    <a href="/">Sourcegraph Console</a>
                </h1>
            </nav>
            <div style={{ flex: 1 }} />
            {data?.user ? <nav className="user">{data.user.email}</nav> : null}
        </div>
    </header>
)
