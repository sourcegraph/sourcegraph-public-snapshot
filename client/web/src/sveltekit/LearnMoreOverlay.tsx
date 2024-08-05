import { FC, useEffect, useRef } from 'react'

import { mdiFilePlusOutline, mdiGraphOutline, mdiMagnifyScan, mdiClose } from '@mdi/js'
import { BrandLogo } from 'src/components/branding/BrandLogo'

import { Button, Icon, ProductStatusBadge } from '@sourcegraph/wildcard'

export const LearnMoreOverlay: FC<{}> = ({}) => {
    const dialogRef = useRef<HTMLDialogElement | null>(null)
    const innerRef = useRef<HTMLDivElement | null>(null)

    useEffect(() => {
        dialogRef.current?.showModal()
    }, [dialogRef, innerRef])

    return (
        <dialog ref={dialogRef}>
            <div className="inner" ref={innerRef}>
                <div className="content">
                    <div className="logo">
                        <BrandLogo variant="symbol" isLightTheme={false} />
                        <ProductStatusBadge status="beta" />
                    </div>
                    <div className="message">
                        <h1>
                            <span>You've activated a better, faster experience</span> ⚡
                        </h1>
                        <p className="subtitle">
                            Get ready for a new Code Search experience: rewritten from the ground-up for performance to
                            empower your workflow.
                        </p>
                    </div>
                    <div className="features">
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
                    <div className="cta">
                        <div>
                            <Button variant="primary" onClick={() => handleDismiss()}>
                                Awesome. I’m ready to use it!
                            </Button>
                        </div>
                        <p> You can opt out at any time by using the toggle at the top of the screen. </p>
                        <p>
                            Whilst exploring the new experience, consider leaving us some feedback via the button at the
                            top. We'd love to hear from you!
                        </p>
                    </div>
                </div>
                {/*

            {#if $isLightTheme}
            <WelcomeOverlayScreenshotLight />
            {:else}
            <WelcomeOverlayScreenshotDark />
            {/if}
            */}

                <Button variant="icon" aria-label="Close welcome overlay" onClick={() => handleDismiss()}>
                    <Icon svgPath={mdiClose} aria-hidden="true" />
                </Button>
            </div>
        </dialog>
    )
}
