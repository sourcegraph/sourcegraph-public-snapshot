import React, { useCallback, useMemo, useState } from 'react'

import { mdiClose, mdiLightningBolt, mdiRadioboxBlank, mdiRadioboxMarked } from '@mdi/js'
import classNames from 'classnames'

import { isMacPlatform } from '@sourcegraph/common'
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

import luckyStyles from './LuckySearchToggle.module.scss'
import styles from './Toggles.module.scss'

interface LuckySearchToggleProps extends Omit<ToggleProps, 'title' | 'iconSvgPath' | 'onToggle'> {
    onSelect: (enabled: boolean) => void
}

/**
 * A toggle displayed in the QueryInput.
 */
export const LuckySearchToggle: React.FunctionComponent<LuckySearchToggleProps> = ({
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

            <LuckySearchToggleMenu onSelect={onSelect} isActive={isActive} closeMenu={() => setIsPopoverOpen(false)} />
        </Popover>
    )
}

const LuckySearchToggleMenu: React.FunctionComponent<
    Pick<LuckySearchToggleProps, 'onSelect' | 'isActive'> & { closeMenu: () => void }
> = ({ onSelect, isActive, closeMenu }) => {
    const onChange = useCallback(
        (value: boolean) => {
            onSelect(value)
            closeMenu()
        },
        [onSelect, closeMenu]
    )

    return (
        <PopoverContent aria-labelledby="lucky-search-popover-header" position={Position.bottomEnd}>
            <div className="d-flex align-items-baseline px-3 py-2">
                <H3 id="lucky-search-popover-header" className="m-0">
                    Smart Search
                </H3>
                <span className="ml-2 text-muted flex-1">{isMacPlatform() ? '⌥' : 'Alt+'}S</span>
                <Button onClick={() => closeMenu()} variant="icon" aria-label="Close">
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <RadioItem
                value={true}
                header="Enabled"
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
    <Label className={luckyStyles.label}>
        <Input
            className="sr-only"
            type="radio"
            name="luckySearch"
            value={value.toString()}
            checked={isChecked}
            onChange={() => onSelect(value)}
        />
        <Icon
            svgPath={isChecked ? mdiRadioboxMarked : mdiRadioboxBlank}
            aria-hidden={true}
            className={classNames(luckyStyles.radioIcon, isChecked && luckyStyles.radioIconActive)}
            inline={false}
        />

        <span className="d-flex flex-column">
            <span className={luckyStyles.radioHeader}>{header}</span>
            <span className={luckyStyles.radioDescription}>{description}</span>
        </span>
    </Label>
)
