import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { View } from './utils/types'

import './NavBar.css'

interface NavBarProps {
	setView: (selectedView: View) => void
	view: View
}

export const NavBar: React.FunctionComponent<React.PropsWithChildren<NavBarProps>> = ({ setView, view }) => (
	<div className="tab-menu-container">
		<VSCodeButton onClick={() => setView('chat')} className="tab-menu-item" appearance="icon" type="button">
			<p className={view === 'chat' ? 'tab-menu-item-selected' : ''}>Ask</p>
		</VSCodeButton>
		<VSCodeButton onClick={() => setView('recipes')} className="tab-menu-item" appearance="icon" type="button">
			<p className={view === 'recipes' ? 'tab-menu-item-selected' : ''}>Recipes</p>
		</VSCodeButton>
		<VSCodeButton onClick={() => setView('about')} className="tab-menu-item" appearance="icon" type="button">
			<p className={view === 'about' ? 'tab-menu-item-selected' : ''}>About</p>
		</VSCodeButton>
		<VSCodeButton onClick={() => setView('settings')} className="tab-menu-item" appearance="icon" type="button">
			<p className={view === 'settings' ? 'tab-menu-item-selected' : ''}>Settings</p>
		</VSCodeButton>
	</div>
)
