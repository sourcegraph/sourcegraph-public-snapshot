import { useState, useEffect } from 'react'

import classNames from 'classnames'

import type { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Text, Button } from '@sourcegraph/wildcard'

import styles from './CodyOnboarding.module.scss'

export function WelcomeStep({
    onNext,
    pro,
    seatCount,
    telemetryRecorder,
}: {
    onNext: () => void
    pro: boolean
    seatCount: number | null
    telemetryRecorder: TelemetryRecorder
}): JSX.Element {
    const [show, setShow] = useState(false)
    const isLightTheme = useIsLightTheme()
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.onboarding.welcome', 'view', {
            metadata: { tier: pro ? 1 : 0, seatCount: seatCount ?? -1 },
        })
    }, [pro, telemetryRecorder])

    useEffect(() => {
        // theme is not ready on first render, it defaults to system theme.
        // so we need to wait a bit before showing the welcome video.
        setTimeout(() => {
            setShow(true)
        }, 500)
    }, [])

    return (
        <div className={classNames('d-flex flex-column align-items-center p-5')}>
            {show ? (
                <>
                    <video width="180" className={classNames('mb-5', styles.welcomeVideo)} autoPlay={true} muted={true}>
                        <source
                            src={
                                isLightTheme
                                    ? 'https://storage.googleapis.com/sourcegraph-assets/hiCodyWhite.mp4'
                                    : 'https://storage.googleapis.com/sourcegraph-assets/hiCodyDark.mp4'
                            }
                            type="video/mp4"
                        />
                        Your browser does not support the video tag.
                    </video>
                    <Text className={classNames('mb-4 pb-4', styles.fadeIn, styles.fadeSecond, styles.welcomeSubtitle)}>
                        Ready to breeze through the basics and get comfortable with Cody
                        {pro ? ' Pro' : ''}?
                    </Text>
                    <Button
                        onClick={onNext}
                        variant="primary"
                        size="lg"
                        className={classNames(styles.fadeIn, styles.fadeThird)}
                    >
                        Sure, let's dive in!
                    </Button>
                </>
            ) : (
                <div className={styles.blankPlaceholder} />
            )}
        </div>
    )
}
