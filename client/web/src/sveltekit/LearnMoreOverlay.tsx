import { useEffect, useRef, forwardRef, useImperativeHandle } from 'react'

import { mdiFilePlusOutline, mdiGraphOutline, mdiMagnifyScan, mdiClose } from '@mdi/js'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, H1, H3, H5, Text, Icon, ProductStatusBadge } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'

import styles from './LearnMoreOverlay.module.scss'

export const LearnMoreOverlay = forwardRef<{ show: () => void; hide: () => void }, { handleEnable: () => void }>(
    ({ handleEnable }, forwardedRef) => {
        const innerRef = useRef<HTMLDivElement | null>(null)
        const dialogRef = useRef<HTMLDialogElement | null>(null)

        const isLightTheme = useIsLightTheme()

        const show = (): void => dialogRef.current?.showModal()
        const hide = (): void => dialogRef.current?.close()

        useImperativeHandle(forwardedRef, () => ({ show, hide }))

        const handleClickOutside = (event: MouseEvent): void => {
            // Use an inner div because the whole backdrop registers as part of the dialog
            if (innerRef.current && !innerRef.current.contains(event.target as Node)) {
                hide()
            }
        }

        useEffect(() => {
            document.body.addEventListener('mousedown', handleClickOutside)
            return () => {
                document.body.removeEventListener('mousedown', handleClickOutside)
            }
        })

        return (
            <dialog ref={dialogRef} className={styles.dialog}>
                <div className={styles.inner} ref={innerRef}>
                    <div className={styles.content}>
                        <div className={styles.logo}>
                            <BrandLogo variant="symbol" isLightTheme={false} disableSymbolSpin={true} />
                            <ProductStatusBadge status="beta" />
                        </div>
                        <div className={styles.message}>
                            <H1>
                                <span>Try a new, faster experience</span>
                            </H1>
                            <Text className={styles.subtitle}>
                                Get ready for a new Code Search experience: rewritten from the ground-up for performance
                                to empower your workflow.
                            </Text>
                        </div>
                        <div className={styles.features}>
                            <div>
                                <Icon svgPath={mdiFilePlusOutline} aria-hidden={true} />
                                <H5>New in-line diff view</H5>
                                <Text>Easily compare commits and see how a file changed over time, all in-line</Text>
                            </div>
                            <div>
                                <Icon svgPath={mdiGraphOutline} aria-hidden={true} />
                                <H5>Revamped code navigation</H5>
                                <Text>
                                    Quickly find a list of references of a given symbol, or immediately jump to the
                                    definition
                                </Text>
                            </div>
                            <div>
                                <Icon svgPath={mdiMagnifyScan} aria-hidden={true} />
                                {/* TODO: add keyboard shortcut */}
                                <H5>Reworked fuzzy finder</H5>
                                <Text>Find files and symbols quickly and easily with our whole new fuzzy finder.</Text>
                            </div>
                        </div>
                        <div className={styles.cta}>
                            <H3>Enable the new UI</H3>
                            <div>
                                <Button variant="primary" onClick={handleEnable}>
                                    Enable
                                </Button>
                                <Button variant="secondary" onClick={hide}>
                                    No thanks
                                </Button>
                            </div>
                            <Text>You can opt out at any time by using the toggle at the top of the screen. </Text>
                            <Text>
                                Whilst exploring the new experience, consider leaving us some feedback via the button at
                                the top. We'd love to hear from you!
                            </Text>
                        </div>
                    </div>

                    <img
                        src={`/.assets/img/welcome-overlay-screenshot-${isLightTheme ? 'light' : 'dark'}.svg`}
                        aria-hidden="true"
                        alt="Example screenshot of beta web app"
                    />

                    <Button variant="icon" aria-label="Close welcome overlay" onClick={hide}>
                        <Icon svgPath={mdiClose} aria-hidden="true" />
                    </Button>
                </div>
            </dialog>
        )
    }
)
