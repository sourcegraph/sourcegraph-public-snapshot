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
	{ tab: 'settings', title: 'Settings' },
]

export const NavBar: React.FunctionComponent<React.PropsWithChildren<NavBarProps>> = ({ setView, view, devMode }) => (
	<div className="tab-menu-container">
		{navBarItems.map(({ title, tab }) => (
			<VSCodeButton onClick={() => setView(tab)} className="tab-menu-item" appearance="icon" type="button">
				<p className={view === tab ? 'tab-menu-item-selected' : ''}>{title}</p>
			</VSCodeButton>
		))}
		{devMode && (
			<VSCodeButton onClick={() => setView('debug')} className="tab-menu-item" appearance="icon" type="button">
				<p className={view === 'debug' ? 'tab-menu-item-selected' : ''}>Debug</p>
			</VSCodeButton>
		)}
	</div>
)
