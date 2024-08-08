import { FC, useRef, useEffect, useCallback } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiChevronDown, mdiClose } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Text, H3, Popover, PopoverTrigger, PopoverContent, Icon, Button } from '@sourcegraph/wildcard'

import { LearnMoreOverlay } from './LearnMoreOverlay'
import { enableSvelteAndReload, canEnableSvelteKit } from './util'

import styles from './SvelteKitNavItem.module.scss'

export const SvelteKitNavItem: FC<{ userID?: string }> = ({ userID }) => {
    const location = useLocation()
    const client = useApolloClient()
    const [departureDismissed, setDepartureDismissed] = useTemporarySetting('webNext.departureMessage.dismissed', false)
    const [showDeparture, _setDepartureDismissed] = useTemporarySetting('webNext.departureMessage.show', false)
    const [_showWelcomeMessage, setShowWelcomeMessage] = useTemporarySetting('webNext.welcomeOverlay.show', false)

    const departureRef = useRef<HTMLDivElement | null>(null)
    const handleClickOutside = useCallback(
        (event: MouseEvent) => {
            if (departureRef.current && !departureRef.current.contains(event.target as Node)) {
                setDepartureDismissed(true)
            }
        },
        [setDepartureDismissed]
    )

    useEffect(() => {
        document.addEventListener('click', handleClickOutside)
        return () => {
            document.removeEventListener('click', handleClickOutside)
        }
    }, [handleClickOutside])

    const learnMoreRef = useRef<{ show: () => void; hide: () => void } | null>(null)

    if (!userID || !canEnableSvelteKit(location.pathname)) {
        return null
    }

    // only show if the welcome message has been dismissed so we know they have been introduced to the new webapp
    const departureVisible = !departureDismissed && showDeparture
    const popoverProps = departureVisible ? { isOpen: true, onOpenChange: () => {} } : {}

    return (
        <>
            <LearnMoreOverlay ref={learnMoreRef} handleEnable={() => enableSvelteAndReload(client, userID)} />
            <Popover {...popoverProps}>
                <PopoverTrigger className={styles.badge}>
                    <Button>
                        Try a new, faster UX
                        <Icon svgPath={mdiChevronDown} aria-hidden="true" />
                    </Button>
                </PopoverTrigger>
                <PopoverContent className={styles.popover} position="bottomEnd">
                    {departureVisible ? (
                        <div ref={departureRef}>
                            <div className={styles.section}>
                                <H3>
                                    <span>Switched out of the new experience?</span>
                                    <Button variant="icon" onClick={() => setDepartureDismissed(true)}>
                                        <Icon svgPath={mdiClose} inline={true} aria-label="close" />
                                    </Button>
                                </H3>
                                <Text>
                                    Remember, you can always switch back using the toggle above. We're still working on
                                    it, so check back soon.
                                </Text>
                            </div>
                            <div className={styles.section}>
                                <Text>Got feedback for us on the beta? We’d love to hear from you.</Text>
                                <Button
                                    as="a"
                                    variant="secondary"
                                    href="https://community.sourcegraph.com/c/code-search/9"
                                    target="_blank"
                                    rel="noreferrer noopener"
                                >
                                    Leave feedback
                                </Button>
                                <Text className={styles.small}>It only takes two minutes and helps a ton!</Text>
                            </div>
                        </div>
                    ) : (
                        <div className={styles.section}>
                            <H3>
                                <span className={styles.colorful}>Try a new, faster UX (Beta)</span>
                                <span className={styles.enableToggle}>
                                    <Toggle
                                        value={false}
                                        onToggle={() => {
                                            setShowWelcomeMessage(true)
                                            enableSvelteAndReload(client, userID)
                                        }}
                                        title="Enable new, faster UX"
                                        className={styles.toggle}
                                    />
                                    Enable
                                </span>
                            </H3>
                            <Text>
                                We've rewritten Code Search from the ground-up for performance to empower your workflow.
                            </Text>
                            <Button variant="secondary" onClick={() => learnMoreRef.current?.show()}>
                                Learn more
                            </Button>
                        </div>
                    )}
                </PopoverContent>
            </Popover>
        </>
    )
}
