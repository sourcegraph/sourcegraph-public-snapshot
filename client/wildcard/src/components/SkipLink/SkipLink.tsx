import React, { useState, useEffect } from 'react'

import VisuallyHidden from '@reach/visually-hidden'

import { AnchorLink } from '../Link'

import styles from './SkipLink.module.scss'

export interface SkipLinksState {
    addLink: (link: string) => void
    removeLink: (link: string) => void
    links: string[]
}
export const SkipLinksContext = React.createContext<SkipLinksState | undefined>(undefined)
SkipLinksContext.displayName = 'SkipLinksContext'

const useSkipLinksContext = (): SkipLinksState => {
    const context = React.useContext(SkipLinksContext)
    if (context === undefined) {
        throw new Error('SkipLink must be used within a SkipLinksProvider')
    }
    return context
}

export const SkipLinksProvider: React.FunctionComponent = ({ children }) => {
    const [links, setLinks] = useState<string[]>([])

    const addLink = (linkToAdd: string): void => {
        setLinks(previous => [...previous, linkToAdd])
    }

    const removeLink = (linkToRemove: string): void => {
        setLinks(previous => previous.filter(link => link !== linkToRemove))
    }

    return (
        <SkipLinksContext.Provider value={{ links, addLink, removeLink }}>
            {links.length > 0 && (
                <div className={styles.skipLinks}>
                    {links.map((link, index) => (
                        <AnchorLink key={link} to={`#${link}`} tabIndex={index + 1}>
                            Link to {link}
                        </AnchorLink>
                    ))}
                </div>
            )}
            {children}
        </SkipLinksContext.Provider>
    )
}

interface SkipLinkProps {
    id: string
    name: string
}

/**
 * Skip links
 */
export const SkipLink: React.FunctionComponent<SkipLinkProps> = ({ id, name }) => {
    const context = useSkipLinksContext()

    useEffect(() => {
        console.log('Adding links!!')
        context.addLink(id)

        return () => {
            context.removeLink(id)
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [id])

    return (
        <VisuallyHidden>
            <AnchorLink to={id} id={id}>
                Hello!
            </AnchorLink>
        </VisuallyHidden>
    )
}
