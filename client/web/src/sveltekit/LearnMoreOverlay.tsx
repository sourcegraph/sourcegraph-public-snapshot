import { useCallback, useEffect, useRef, forwardRef, useImperativeHandle } from 'react'

import { mdiFilePlusOutline, mdiGraphOutline, mdiMagnifyScan, mdiClose } from '@mdi/js'
import { BrandLogo } from 'src/components/branding/BrandLogo'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, Icon, ProductStatusBadge } from '@sourcegraph/wildcard'

import styles from './LearnMoreOverlay.module.scss'

export const LearnMoreOverlay = forwardRef<_, { handleEnable: () => void }>(({ handleEnable }, forwardedRef) => {
    const innerRef = useRef<HTMLDivElement | null>(null)
    const dialogRef = useRef<HTMLDialogElement | null>(null)

    const isLightTheme = useIsLightTheme()

    const show = () => dialogRef.current?.showModal()
    const hide = () => dialogRef.current?.close()

    useImperativeHandle(forwardedRef, () => ({ show, hide }))

    const handleClickOutside = (event: MouseEvent) => {
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
                        <h1>
                            <span>Try a new, faster experience</span>
                        </h1>
                        <p className={styles.subtitle}>
                            Get ready for a new Code Search experience: rewritten from the ground-up for performance to
                            empower your workflow.
                        </p>
                    </div>
                    <div className={styles.features}>
                        <div>
                            <Icon svgPath={mdiFilePlusOutline} aria-hidden={true} />
                            <h5>New in-line diff view</h5>
                            <p>Easily compare commits and see how a file changed over time, all in-line</p>
                        </div>
                        <div>
                            <Icon svgPath={mdiGraphOutline} aria-hidden={true} />
                            <h5>Revamped code navigation</h5>
                            <p>
                                Quickly find a list of references of a given symbol, or immediately jump to the
                                definition
                            </p>
                        </div>
                        <div>
                            <Icon svgPath={mdiMagnifyScan} aria-hidden={true} />
                            {/* TODO: add keyboard shortcut */}
                            <h5>Reworked fuzzy finder</h5>
                            <p>Find files and symbols quickly and easily with our whole new fuzzy finder.</p>
                        </div>
                    </div>
                    <div className={styles.cta}>
                        <div>
                            <Button variant="primary" onClick={handleEnable}>
                                Enable
                            </Button>
                            <Button variant="secondary" onClick={hide}>
                                No thanks
                            </Button>
                        </div>
                        <p> You can opt out at any time by using the toggle at the top of the screen. </p>
                        <p>
                            Whilst exploring the new experience, consider leaving us some feedback via the button at the
                            top. We'd love to hear from you!
                        </p>
                    </div>
                </div>

                <img
                    src={`/.assets/img/welcome-overlay-screenshot-${isLightTheme ? 'light' : 'dark'}.svg`}
                    aria-hidden="true"
                />

                <Button variant="icon" aria-label="Close welcome overlay" onClick={hide}>
                    <Icon svgPath={mdiClose} aria-hidden="true" />
                </Button>
            </div>
        </dialog>
    )
})
