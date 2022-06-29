import React, { useState } from 'react'

import { Button } from '../Button'
import { AnchorLink } from '../Link'

import { SkipLinkState } from './SkipLink'

import styles from './SkipLink.module.scss'

interface SkipLinkContextState {
    addLink: (link: SkipLinkState) => void
    removeLink: (id: SkipLinkState['id']) => void
    links: SkipLinkState[]
}

const SkipLinkContext = React.createContext<SkipLinkContextState | undefined>(undefined)
SkipLinkContext.displayName = 'SkipLinkContext'

export const useSkipLinkContext = (): SkipLinkContextState => {
    const context = React.useContext(SkipLinkContext)
    if (context === undefined) {
        throw new Error('SkipLink must be used within a SkipLinkProvider')
    }
    return context
}

export const SkipLinkProvider: React.FunctionComponent = ({ children }) => {
    const [links, setLinks] = useState<SkipLinkState[]>([])

    const addLink = (linkToAdd: SkipLinkState): void => {
        setLinks(previous => [...previous, linkToAdd])
    }

    const removeLink = (linkIdToRemove: SkipLinkState['id']): void => {
        setLinks(previous => previous.filter(link => link.id !== linkIdToRemove))
    }

    return (
        <SkipLinkContext.Provider value={{ links, addLink, removeLink }}>
            {links.length > 0 && (
                <nav>
                    <ul className={styles.list}>
                        {links.map((link, index) => (
                            <li key={link.id}>
                                <Button variant="secondary" as={AnchorLink} to={`#${link.id}`} className={styles.link}>
                                    Skip to {link.name}
                                </Button>
                            </li>
                        ))}
                    </ul>
                </nav>
            )}
            {children}
        </SkipLinkContext.Provider>
    )
}
