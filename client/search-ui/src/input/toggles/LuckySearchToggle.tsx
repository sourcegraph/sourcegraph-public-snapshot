import React, { useCallback, useMemo, useState } from 'react'

import { mdiClose, mdiLightningBolt } from '@mdi/js'
import classNames from 'classnames'

import { isMacPlatform } from '@sourcegraph/common'
import { Button, H3, Icon, Input, Label, Popover, PopoverContent, PopoverTrigger, Tooltip } from '@sourcegraph/wildcard'

import { ToggleProps } from './QueryInputToggle'

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
            <PopoverContent>
                <LuckySearchToggleMenu
                    onSelect={onSelect}
                    isActive={isActive}
                    closeMenu={() => setIsPopoverOpen(false)}
                />
            </PopoverContent>
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
        <div>
            <div>
                <H3>Smart Search ({isMacPlatform() ? '⌥' : 'Alt+'}S) </H3>
                <Button onClick={() => closeMenu()} variant="icon" aria-label="Close">
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <Label>
                <Input
                    type="radio"
                    name="luckySearch"
                    value="true"
                    checked={isActive}
                    onChange={() => onChange(true)}
                />
                <span>Enable</span>
                <span>Suggest variations of your query to find more results that may relate.</span>
            </Label>
            <Label>
                <Input
                    type="radio"
                    name="luckySearch"
                    value="false"
                    checked={!isActive}
                    onChange={() => onChange(false)}
                />
                <span>Disable</span>
                <span>Only show results that precisely match your query.</span>
            </Label>
        </div>
    )
}
