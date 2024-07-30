import { FC, useRef, useEffect } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiHelpCircleOutline, mdiClose } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Text, H3, Popover, PopoverTrigger, PopoverContent, Icon, Button, Link } from '@sourcegraph/wildcard'

import { enableSvelteAndReload, canEnableSvelteKit } from './util'

import styles from './SvelteKitNavItem.module.scss'

export const SvelteKitNavItem: FC<{ userID?: string }> = ({ userID }) => {
    const location = useLocation()
    const client = useApolloClient()
    const [toggledOff] = useTemporarySetting('webNext.toggled.off', false)
    const [departureDismissed, setDepartureDismissed] = useTemporarySetting('webNext.departureMessage.dismissed', false)
    const [_toggledOn, setToggledOn] = useTemporarySetting('webNext.toggled.on', false)

    if (!userID || !canEnableSvelteKit(location.pathname)) {
        return null
    }

    let departureRef = useRef(null)

    const handleClickOutside = event => {
        console.log('handling')
        console.log(departureRef.current)
        if (departureRef.current && !departureRef.current.contains(event.target)) {
            console.log('dismissing')
            setDepartureDismissed(true)
        }
    }

    useEffect(() => {
        document.addEventListener('click', handleClickOutside)
        return () => {
            document.removeEventListener('click', handleClickOutside)
        }
    }, [])

    const showDeparture = toggledOff && !departureDismissed

    return (
        <Popover isOpen={showDeparture ? true : undefined}>
            <PopoverTrigger className={styles.badge}>
                <div className={styles.container}>
                    <Icon svgPath={mdiHelpCircleOutline} />
                    <Text>New, faster UX</Text>
                    <Toggle
                        value={false}
                        onToggle={() => {
                            setToggledOn(true)
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
                                    <Icon svgPath={mdiClose} inline />
                                </Button>
                            </H3>
                            <Text>
                                Remember, you can always switch back using the toggle above. It’ll keep getting improved
                                until full release, so check back soon.
                            </Text>
                        </div>
                        <div className={styles.section}>
                            <Text>Got feedback for us on the beta? We’d love to hear from you.</Text>
                            <Button
                                as="a"
                                variant="secondary"
                                href="https://community.sourcegraph.com/"
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
                            performance, which now available in beta.
                        </Text>
                        <Text>Simply activate the toggle to get it.</Text>
                    </div>
                )}
            </PopoverContent>
        </Popover>
    )
}
