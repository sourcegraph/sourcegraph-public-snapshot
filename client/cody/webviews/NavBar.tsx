import React from 'react'

import styles from './NavBar.module.css'

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
    <div className={styles.tabMenuContainer}>
        <div className={styles.tabMenuGroup}>
            {navBarItems.map(({ title, tab }) => (
                <button key={title} onClick={() => setView(tab)} className={styles.tabBtn} type="button">
                    <span className={view === tab ? styles.tabBtnSelected : ''}>{title}</span>
                </button>
            ))}
            {devMode && (
                <button onClick={() => setView('debug')} className={styles.tabBtn} type="button">
                    <span className={view === 'debug' ? styles.tabBtnSelected : ''}>Debug</span>
                </button>
            )}
        </div>
    </div>
)
