import classNames from 'classnames'
import React, { useCallback, useEffect, useRef, useMemo } from 'react'
import { fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'
import { Key } from 'ts-key-enum'

import { PatternTypeProps, CaseSensitivityProps } from '../..'

export interface ToggleProps extends PatternTypeProps, CaseSensitivityProps {
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
}

/**
 * A toggle displayed in the QueryInput.
 */
export const QueryInputToggle: React.FunctionComponent<ToggleProps> = ({ onToggle, ...props }) => {
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

    return (
        // Click events here are defined in useEffect
        // eslint-disable-next-line jsx-a11y/click-events-have-key-events
        <div
            ref={toggleCheckbox}
            onClick={onCheckboxToggled}
            className={classNames(
                'btn btn-icon toggle-container__toggle',
                props.className,
                { disabled: !!disabledRule },
                { 'toggle-container__toggle--active': isActive },
                props.activeClassName
            )}
            role="checkbox"
            aria-disabled={!!disabledRule}
            aria-checked={isActive}
            aria-label={`${props.title} toggle`}
            tabIndex={0}
            data-tooltip={tooltipValue}
        >
            <Icon className="icon-inline" />
        </div>
    )
}
