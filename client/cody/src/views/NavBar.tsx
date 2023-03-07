import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

interface NavBarProps {
    setView: (selectedView: number) => void
    onLogoutClick: () => void
    view: number
}

export const NavBar: React.FunctionComponent<React.PropsWithChildren<NavBarProps>> = ({
    setView,
    view,
    onLogoutClick,
}) => (
    <div className="tab-menu-container">
        <VSCodeButton
            onClick={() => setView(1)}
            id="tab-menu-item-ask"
            className="tab-menu-item"
            appearance="icon"
            type="button"
        >
            <p className={view === 1 ? 'tab-menu-item-selected' : ''}>Ask</p>
        </VSCodeButton>
        <VSCodeButton
            onClick={() => setView(2)}
            id="tab-menu-item-recipes"
            className="tab-menu-item"
            appearance="icon"
            type="button"
        >
            <p className={view === 2 ? 'tab-menu-item-selected' : ''}>Recipes</p>
        </VSCodeButton>
        <VSCodeButton
            onClick={() => setView(3)}
            id="tab-menu-item-debug"
            className="tab-menu-item"
            appearance="icon"
            type="button"
        >
            <p className={view === 3 ? 'tab-menu-item-selected' : ''}>About</p>
        </VSCodeButton>
        <VSCodeButton
            onClick={() => onLogoutClick()}
            id="tab-menu-item-logout"
            className="tab-menu-item"
            appearance="icon"
            type="button"
        >
            <p>Logout</p>
        </VSCodeButton>
    </div>
)
