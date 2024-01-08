import React, { useCallback, useEffect, useState } from 'react'

import { mdiClose, mdiRadioboxBlank, mdiRadioboxMarked, mdiHeart } from '@mdi/js'
import classNames from 'classnames'

import {
    Button,
    Icon,
    Input,
    Label,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    Tooltip,
    H4,
    H2,
} from '@sourcegraph/wildcard'

import type { ToggleProps } from './QueryInputToggle'

import smartStyles from './SmartSearchToggle.module.scss'
import styles from './Toggles.module.scss'

export const smartSearchIconSvgPath =
    'M11.3956 20H10.2961L11.3956 13.7778H7.54754C6.58003 13.7778 7.18473 13.1111 7.20671 13.0844C8.62499 11.0578 10.7579 8.03556 13.6054 4H14.7049L13.6054 10.2222H17.4645C17.9042 10.2222 18.1461 10.3911 17.9042 10.8089C13.5615 16.9333 11.3956 20 11.3956 20Z'

export enum SearchModes {
    Smart = 'Smart',
    PreciseNew = 'Precise (NEW) 💖',
    Precise = 'Precise (legacy)',
}

interface SmartSearchToggleProps extends Omit<ToggleProps, 'title' | 'iconSvgPath' | 'onToggle' | 'isActive'> {
    onSelect: (mode: SearchModes) => void
    mode: SearchModes
}

/**
 * A toggle displayed in the QueryInput.
 */
export const SmartSearchToggleExtended: React.FunctionComponent<SmartSearchToggleProps> = ({
    onSelect,
    interactive = true,
    mode,
    className,
}) => {
    const tooltipValue = mode.toString()

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
                        !interactive && styles.toggleNonInteractive,
                        mode !== SearchModes.Precise && styles.toggleActive
                    )}
                    variant="icon"
                    {...interactiveProps}
                >
                    <Icon
                        aria-label={tooltipValue}
                        svgPath={mode === SearchModes.PreciseNew ? mdiHeart : smartSearchIconSvgPath}
                        // compensate for left margin set on "svg" for the flash symbol
                        className={mode === SearchModes.PreciseNew ? 'ml-0' : ''}
                    />
                </PopoverTrigger>
            </Tooltip>

            <SmartSearchToggleMenu onSelect={onSelect} mode={mode} closeMenu={() => setIsPopoverOpen(false)} />
        </Popover>
    )
}

const SmartSearchToggleMenu: React.FunctionComponent<
    Pick<SmartSearchToggleProps, 'onSelect' | 'mode'> & { closeMenu: () => void }
> = ({ onSelect, mode, closeMenu }) => {
    const [getMode, setMode] = useState(mode)
    useEffect(() => {
        setMode(mode)
    }, [mode])

    const onChange = useCallback(
        (value: SearchModes) => {
            setMode(value)
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
            aria-labelledby="smart-search-popover-header"
            position={Position.bottomEnd}
            className={smartStyles.popoverWindow}
        >
            <div className="d-flex justify-content-end px-3 py-2">
                <H4 as={H2} id="smart-search-popover-header" className="m-0 flex-1">
                    Search Mode Picker
                </H4>
                <Button onClick={() => closeMenu()} variant="icon" aria-label="Close">
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <RadioItem
                value={SearchModes.Smart}
                header="Smart"
                description="Suggest variations of your query to find more results that may relate."
                isChecked={getMode === SearchModes.Smart}
                onSelect={onChange}
            />
            <RadioItem
                value={SearchModes.PreciseNew}
                header="Precise (NEW) 💖"
                description='Spaces are interpreted as AND. "repo" and "file" filters expect glob syntax.'
                isChecked={getMode === SearchModes.PreciseNew}
                onSelect={onChange}
            />
            <RadioItem
                value={SearchModes.Precise}
                header="Precise (legacy)"
                description='Spaces are interpreted literaly. "repo" and "file" filters expect regex syntax.'
                isChecked={getMode === SearchModes.Precise}
                onSelect={onChange}
            />
        </PopoverContent>
    )
}

const RadioItem: React.FunctionComponent<{
    value: SearchModes
    isChecked: boolean
    onSelect: (value: SearchModes) => void
    header: string
    description: string
}> = ({ value, isChecked, onSelect, header, description }) => (
    <Label className={smartStyles.label}>
        <Input
            className="sr-only"
            type="radio"
            name="smartSearch"
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
