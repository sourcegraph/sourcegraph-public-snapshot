import { FC, useRef, useEffect, useCallback } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiHelpCircleOutline, mdiClose } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Text, H3, Popover, PopoverTrigger, PopoverContent, Icon, Button } from '@sourcegraph/wildcard'

import { enableSvelteAndReload, canEnableSvelteKit } from './util'

import styles from './SvelteKitNavItem.module.scss'

export const SvelteKitNavItem: FC<{ userID?: string }> = ({ userID }) => {
    const location = useLocation()
    const client = useApolloClient()
    const [departureDismissed, setDepartureDismissed] = useTemporarySetting('webNext.departureMessage.dismissed', false)
    const [_welcomeDismissed, setWelcomeDismissed] = useTemporarySetting('webNext.welcomeOverlay.dismissed', false)

    const departureRef = useRef<HTMLDivElement | null>(null)

    const handleClickOutside = useCallback(
        (event: MouseEvent) => {
            if (departureRef.current && !departureRef.current.contains(event.target as Node)) {
                setDepartureDismissed(true)
            }
        },
        [departureRef, setDepartureDismissed]
    )

    useEffect(() => {
        document.addEventListener('click', handleClickOutside)
        return () => {
            document.removeEventListener('click', handleClickOutside)
        }
    }, [handleClickOutside])

    if (!userID || !canEnableSvelteKit(location.pathname)) {
        return null
    }

    const showDeparture = !departureDismissed
    const popoverProps = showDeparture ? { isOpen: true, onOpenChange: () => {} } : {}

    return (
        <Popover {...popoverProps}>
            <PopoverTrigger className={styles.badge}>
                <div className={styles.container}>
                    <Icon className={styles.helpIcon} svgPath={mdiHelpCircleOutline} aria-hidden={true} />
                    <Text>New, faster UX</Text>
                    <Toggle
                        value={false}
                        onToggle={() => {
                            setWelcomeDismissed(false) // Show welcome after switching on
                            enableSvelteAndReload(client, userID)
                        }}
                        title="Enable new, faster UX"
                        className={styles.toggle}
                    />
                </div>
            </PopoverTrigger>
            <PopoverContent className={styles.popover} position="bottomEnd">
                {showDeparture ? (
                    <div ref={departureRef}>
                        <div className={styles.section}>
                            <H3>
                                <span>Switched out of the new experience?</span>
                                <Button variant="icon" onClick={() => setDepartureDismissed(true)}>
                                    <Icon svgPath={mdiClose} inline={true} aria-label="close" />
                                </Button>
                            </H3>
                            <Text>
                                Remember, you can always switch back using the toggle above. We're still working on it,
                                so check back soon.
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
                        <H3>What's this "New, faster UX"?</H3>
                        <Text>
                            We've been busy at work on a new Code Search experience, built from the ground up for
                            performance, which is now available in beta.
                        </Text>
                        <Text>Simply activate the toggle to get it.</Text>
                    </div>
                )}
            </PopoverContent>
        </Popover>
    )
}
