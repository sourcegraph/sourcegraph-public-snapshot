import classNames from 'classnames'
import React, { useCallback, useEffect, useRef, useMemo } from 'react'
import { fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'
import { Key } from 'ts-key-enum'

import styles from './Toggles.module.scss'

export interface ToggleProps {
    /** Title of the toggle.  */
    title: string
    /** Icon to display.  */
    icon: React.ComponentType<{ className?: string }>
    /** Condition for when the toggle should have an active state.  */
    isActive: boolean
    /** Callback on toggle.  */
    onToggle: () => void
    /**
     * A list of conditions to disable the toggle, displaying an associated tooltip when the condition is true.
     * For multiple true conditions, use the first rule that evalutes to true.
     */
    disableOn?: { condition: boolean; reason: string }[]
    className?: string
    activeClassName?: string
    /**
     * If set to false makes the button non-actionable. The main use case for
     * this prop is showing the toggles in examples. This is different from
     * being disabled, because the buttons still render normally.
     */
    interactive?: boolean
}

/**
 * A toggle displayed in the QueryInput.
 */
export const QueryInputToggle: React.FunctionComponent<ToggleProps> = ({ onToggle, interactive = true, ...props }) => {
    const toggleCheckbox = useRef<HTMLDivElement | null>(null)

    const disabledRule = useMemo(() => props.disableOn?.find(({ condition }) => condition), [props.disableOn])
    const onCheckboxToggled = useCallback(() => {
        if (disabledRule) {
            return
        }
        onToggle()
    }, [disabledRule, onToggle])
    const tooltipValue = useMemo(
        () => disabledRule?.reason ?? `${props.isActive ? 'Disable' : 'Enable'} ${props.title.toLowerCase()}`,
        [disabledRule, props.isActive, props.title]
    )
    useEffect(() => {
        const subscription = fromEvent<KeyboardEvent>(window, 'keydown')
            .pipe(
                filter(
                    event =>
                        document.activeElement === toggleCheckbox.current &&
                        (event.key === Key.Enter || event.key === ' ')
                )
            )
            .subscribe(event => {
                event.preventDefault()
                onCheckboxToggled()
            })
        return () => subscription.unsubscribe()
    }, [onCheckboxToggled])

    const Icon = props.icon
    const isActive = props.isActive && !disabledRule

    const interactiveProps = interactive
        ? { tabIndex: 0, 'data-tooltip': tooltipValue, onClick: onCheckboxToggled }
        : {}

    return (
        // Click events here are defined in useEffect
        // eslint-disable-next-line jsx-a11y/click-events-have-key-events
        <div
            ref={toggleCheckbox}
            className={classNames(
                'btn btn-icon',
                styles.toggle,
                props.className,
                !!disabledRule && styles.disabled,
                isActive && styles.toggleActive,
                !interactive && styles.toggleNonInteractive,
                props.activeClassName
            )}
            role="checkbox"
            aria-disabled={!!disabledRule}
            aria-checked={isActive}
            aria-label={`${props.title} toggle`}
            {...interactiveProps}
        >
            <Icon className="icon-inline" />
        </div>
    )
}
