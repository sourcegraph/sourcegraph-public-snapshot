import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiClose, mdiGoogle, mdiHeart, mdiRadioboxBlank, mdiRadioboxMarked } from '@mdi/js'
import classNames from 'classnames'

import {
    Button,
    Icon,
    Input,
    Label,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Tooltip,
    Position,
    H4,
    H2,
} from '@sourcegraph/wildcard'

import type { ToggleProps } from './QueryInputToggle'

import smartStyles from './SmartSearchToggle.module.scss'
import styles from './Toggles.module.scss'

interface FlexibleSearchToggleProps extends Omit<ToggleProps, 'title' | 'iconSvgPath' | 'onToggle'> {
    onSelect: (enabled: boolean) => void
}

/**
 * A toggle displayed in the QueryInput.
 */
export const FlexibleSearchToggle: React.FunctionComponent<FlexibleSearchToggleProps> = ({
    onSelect,
    interactive = true,
    isActive,
    className,
    disableOn,
}) => {
    const disabledRule = useMemo(() => disableOn?.find(({ condition }) => condition), [disableOn])
    const tooltipValue = useMemo(
        () => (disabledRule?.reason ?? isActive ? 'Flexible Search enabled' : 'Flexible Search disabled'),
        [disabledRule?.reason, isActive]
    )

    const interactiveProps = interactive ? {} : { tabIndex: -1, 'aria-hidden': true }

    const [isPopoverOpen, setIsPopoverOpen] = useState(false)

    return (
        <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
            <Tooltip content={tooltipValue} placement="bottom">
                <PopoverTrigger
                    as={Button}
                    className={classNames(
                        'a11y-ignore',
                        styles.toggle,
                        smartStyles.button,
                        className,
                        !!disabledRule && styles.disabled,
                        isActive && styles.toggleActive,
                        !interactive && styles.toggleNonInteractive
                    )}
                    variant="icon"
                    aria-checked={isActive}
                    {...interactiveProps}
                >
                    <Icon aria-label={tooltipValue} svgPath={mdiHeart} />
                </PopoverTrigger>
            </Tooltip>

            <FlexibleSearchToggleMenu
                onSelect={onSelect}
                isActive={isActive}
                closeMenu={() => setIsPopoverOpen(false)}
            />
        </Popover>
    )
}

const FlexibleSearchToggleMenu: React.FunctionComponent<
    Pick<FlexibleSearchToggleProps, 'onSelect' | 'isActive'> & { closeMenu: () => void }
> = ({ onSelect, isActive, closeMenu }) => {
    const [visibleIsEnabled, setVisibleIsEnabled] = useState(isActive)
    useEffect(() => {
        setVisibleIsEnabled(isActive)
    }, [isActive])

    const onChange = useCallback(
        (value: boolean) => {
            setVisibleIsEnabled(value)
            // Wait a tiny bit for user to see the selection change before closing the popover
            setTimeout(() => {
                onSelect(value)
                closeMenu()
            }, 100)
        },
        [onSelect, closeMenu]
    )

    return (
        <PopoverContent
            aria-labelledby="flexible-search-popover-header"
            position={Position.bottomEnd}
            className={smartStyles.popoverWindow}
        >
            <div className="d-flex align-items-center px-3 py-2">
                <H4 as={H2} id="flexible-search-popover-header" className="m-0 flex-1">
                    Flexible Search
                </H4>
                <Button onClick={() => closeMenu()} variant="icon" aria-label="Close">
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <RadioItem
                value={true}
                header="Enable"
                description="Suggest variations of your query to find more results that may relate."
                isChecked={visibleIsEnabled}
                onSelect={onChange}
            />
            <RadioItem
                value={false}
                header="Disable"
                description="Only show results that precisely match your query."
                isChecked={!visibleIsEnabled}
                onSelect={onChange}
            />
        </PopoverContent>
    )
}

const RadioItem: React.FunctionComponent<{
    value: boolean
    isChecked: boolean
    onSelect: (value: boolean) => void
    header: string
    description: string
}> = ({ value, isChecked, onSelect, header, description }) => (
    <Label className={smartStyles.label}>
        <Input
            className="sr-only"
            type="radio"
            name="flexibleSearch"
            value={value.toString()}
            checked={isChecked}
            onChange={() => onSelect(value)}
        />
        <Icon
            svgPath={isChecked ? mdiRadioboxMarked : mdiRadioboxBlank}
            aria-hidden={true}
            className={classNames(smartStyles.radioIcon, isChecked && smartStyles.radioIconActive)}
            inline={false}
        />

        <span className="d-flex flex-column">
            <span className={smartStyles.radioHeader}>{header}</span>
            <span className={smartStyles.radioDescription}>{description}</span>
        </span>
    </Label>
)
