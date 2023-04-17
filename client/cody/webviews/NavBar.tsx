import './NavBar.css'

export type View = 'chat' | 'recipes' | 'login' | 'settings' | 'debug' | 'history'

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
    { tab: 'chat', title: 'Chat' },
    { tab: 'recipes', title: 'Recipes' },
]

export const NavBar: React.FunctionComponent<React.PropsWithChildren<NavBarProps>> = ({ setView, view, devMode }) => (
    <div className="tab-menu-container">
        <div className="tab-menu-group">
            {navBarItems.map(({ title, tab }) => (
                <button key={title} onClick={() => setView(tab)} className="tab-btn" type="button">
                    <p className={view === tab ? 'tab-btn-selected' : ''}>{title}</p>
                </button>
            ))}
            {devMode && (
                <button onClick={() => setView('debug')} className="tab-btn" type="button">
                    <p className={view === 'debug' ? 'tab-btn-selected' : ''}>Debug</p>
                </button>
            )}
        </div>
        <div className="tab-menu-group">
            <button onClick={() => setView('settings')} className="tab-btn" type="button" title="Settings">
                <i
                    className={
                        view === 'settings'
                            ? 'codicon codicon-three-bars tab-btn-selected'
                            : 'codicon codicon-three-bars'
                    }
                />
            </button>
        </div>
    </div>
)
