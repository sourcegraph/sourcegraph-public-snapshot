import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { View } from './utils/types'

import './NavBar.css'

interface NavBarProps {
    setView: (selectedView: View) => void
    view: View
    devMode: boolean
}

interface NavBarItem {
    title: string
    tab: View
}

const navBarItems: NavBarItem[] = [
    { tab: 'chat', title: 'Ask' },
    { tab: 'recipes', title: 'Recipes' },
]

export const NavBar: React.FunctionComponent<React.PropsWithChildren<NavBarProps>> = ({ setView, view, devMode }) => (
    <div className="tab-menu-container">
        <div className="tab-menu-group">
            {navBarItems.map(({ title, tab }) => (
                <VSCodeButton onClick={() => setView(tab)} className="tab-menu-item" appearance="icon" type="button">
                    <p className={view === tab ? 'tab-menu-item-selected' : ''}>{title}</p>
                </VSCodeButton>
            ))}
            {devMode && (
                <VSCodeButton
                    onClick={() => setView('debug')}
                    className="tab-menu-item"
                    appearance="icon"
                    type="button"
                >
                    <p className={view === 'debug' ? 'tab-menu-item-selected' : ''}>Debug</p>
                </VSCodeButton>
            )}
        </div>
        <div className="tab-menu-group">
            <VSCodeButton
                onClick={() => setView('settings')}
                className="tab-menu-item"
                appearance="icon"
                type="button"
                title="Settings"
            >
                <p className={view === 'settings' ? 'tab-menu-item-selected' : ''}>
                    <i className="codicon codicon-three-bars" />
                </p>
            </VSCodeButton>
        </div>
    </div>
)
