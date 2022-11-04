import React, { useCallback, useMemo, useState } from 'react'

import { mdiClose, mdiLightningBolt, mdiRadioboxBlank, mdiRadioboxMarked } from '@mdi/js'
import classNames from 'classnames'

import {
    Button,
    H3,
    Icon,
    Input,
    Label,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Tooltip,
    Position,
} from '@sourcegraph/wildcard'

import { ToggleProps } from './QueryInputToggle'

import smartStyles from './SmartSearchToggle.module.scss'
import styles from './Toggles.module.scss'

interface SmartSearchToggleProps extends Omit<ToggleProps, 'title' | 'iconSvgPath' | 'onToggle'> {
    onSelect: (enabled: boolean) => void
}

/**
 * A toggle displayed in the QueryInput.
 */
export const SmartSearchToggle: React.FunctionComponent<SmartSearchToggleProps> = ({
    onSelect,
    interactive = true,
    isActive,
    className,
    disableOn,
}) => {
    const disabledRule = useMemo(() => disableOn?.find(({ condition }) => condition), [disableOn])
    const tooltipValue = useMemo(
        () => (disabledRule?.reason ?? isActive ? 'Smart Search enabled' : 'Smart Search disabled'),
        [disabledRule?.reason, isActive]
    )

    const interactiveProps = interactive ? {} : { tabIndex: -1, 'aria-hidden': true }

    const [isPopoverOpen, setIsPopoverOpen] = useState(false)

    return (
        <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
            <PopoverTrigger
                as={Button}
                className={classNames(
                    styles.toggle,
                    className,
                    !!disabledRule && styles.disabled,
                    isActive && styles.toggleActive,
                    !interactive && styles.toggleNonInteractive
                )}
                variant="icon"
                aria-checked={isActive}
                {...interactiveProps}
            >
                <Tooltip content={tooltipValue} placement="bottom">
                    <Icon aria-label={tooltipValue} svgPath={mdiLightningBolt} />
                </Tooltip>
            </PopoverTrigger>

            <SmartSearchToggleMenu onSelect={onSelect} isActive={isActive} closeMenu={() => setIsPopoverOpen(false)} />
        </Popover>
    )
}

const SmartSearchToggleMenu: React.FunctionComponent<
    Pick<SmartSearchToggleProps, 'onSelect' | 'isActive'> & { closeMenu: () => void }
> = ({ onSelect, isActive, closeMenu }) => {
    const onChange = useCallback(
        (value: boolean) => {
            onSelect(value)
            closeMenu()
        },
        [onSelect, closeMenu]
    )

    return (
        <PopoverContent
            aria-labelledby="smart-search-popover-header"
            position={Position.bottomEnd}
            className={smartStyles.popoverWindow}
        >
            <div className="d-flex align-items-center px-3 py-2">
                <H3 id="smart-search-popover-header" className="m-0 flex-1">
                    Smart Search
                </H3>
                <Button onClick={() => closeMenu()} variant="icon" aria-label="Close">
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <RadioItem
                value={true}
                header="Enable"
                description="Suggest variations of your query to find more results that may relate."
                isChecked={isActive}
                onSelect={onChange}
            />
            <RadioItem
                value={false}
                header="Disable"
                description="Only show results that precisely match your query."
                isChecked={!isActive}
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
