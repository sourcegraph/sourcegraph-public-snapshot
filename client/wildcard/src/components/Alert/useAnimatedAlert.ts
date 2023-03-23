import { useCallback, useEffect, useRef, useState } from 'react'

import { Assertiveness } from '@react-aria/live-announcer'
import { SpringValue, useSpring } from 'react-spring'

import { useStopwatch } from '../../hooks'
import { screenReaderAnnounce } from '../../utils'

interface AnimatedAlertOptions {
    ariaAnnouncement: {
        /** The message to announce to screen readers when the alert is shown. */
        message: string
        /** The politeness level of the announcement. Defaults 'polite'. */
        politeness?: Assertiveness
    }
    /** Whether the alert starts visible or not. Defaults false. */
    defaultShown?: boolean
    /**
     * How long the alert should be visible for before auto-hiding. Leave out to manually
     * control dismissal.
     */
    autoDuration?: 'short' | 'long'
}

interface AnimatedAlertControls {
    /** Whether or not the alert is currently visible. */
    isShown: boolean
    /** Callback to show the alert. */
    show: () => void
    /** Callback to manually dismiss the alert. */
    dismiss: () => void
    /** Ref to the alert element. */
    ref: React.RefObject<HTMLDivElement>
    /** Style object to be passed to the react-spring animated element. */
    style: {
        height: SpringValue<string>
        opacity: SpringValue<number>
    }
}

// The duration in seconds the alert should be shown before it is automatically hidden
const DURATION_SHORT_S = 4
const DURATION_LONG_S = 10
// Used as a default in case the Alert is not yet rendered. Is the approximate height of a
// standard Alert with a single line of text.
const APPROX_MIN_BANNER_HEIGHT_PX = 40

/**
 * Custom hook to show and hide an accessible alert with an animated transition. The alert
 * can be controlled with the callback functions returned by this hook, or it can be
 * automatically hidden after a certain duration by passing an autoDuration option.
 *
 * @param opts any AnimatedAlertOptions to apply to the controls for this alert
 * @returns the AnimatedAlertControls for this alert
 */
export const useAnimatedAlert = ({
    defaultShown = false,
    autoDuration,
    ariaAnnouncement,
}: AnimatedAlertOptions): AnimatedAlertControls => {
    const [showAlert, setShowAlert] = useState<boolean>(defaultShown)

    // Use a stopwatch to show the alert for a certain duration.
    const {
        time: { seconds },
        start: startTimer,
        stop: stopTimer,
        isRunning,
    } = useStopwatch(defaultShown)

    // Automatically hide the alert after a certain duration if autoDuration is set.
    useEffect(() => {
        if (isRunning && autoDuration && seconds > (autoDuration === 'short' ? DURATION_SHORT_S : DURATION_LONG_S)) {
            stopTimer()
            setShowAlert(false)
        }
    }, [isRunning, stopTimer, seconds, autoDuration])

    const show = useCallback(() => {
        screenReaderAnnounce(ariaAnnouncement.message, ariaAnnouncement.politeness)
        setShowAlert(true)
        startTimer()
    }, [startTimer, ariaAnnouncement.message, ariaAnnouncement.politeness])

    const dismiss = useCallback(() => {
        setShowAlert(false)
        stopTimer()
    }, [stopTimer])

    const ref = useRef<HTMLDivElement>(null)
    let height = ref.current?.offsetHeight || APPROX_MIN_BANNER_HEIGHT_PX
    if (ref.current) {
        height += parseInt(window.getComputedStyle(ref.current).getPropertyValue('margin-top'), 10)
        height += parseInt(window.getComputedStyle(ref.current).getPropertyValue('margin-bottom'), 10)
    }

    const style = useSpring({
        height: showAlert ? `${height}px` : '0px',
        opacity: showAlert ? 1 : 0,
    })

    return {
        isShown: showAlert,
        show,
        dismiss,
        ref,
        style,
    }
}
