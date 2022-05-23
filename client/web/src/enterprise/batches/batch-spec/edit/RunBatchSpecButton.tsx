import React, { useState } from 'react'

import VisuallyHidden from '@reach/visually-hidden'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import { animated } from 'react-spring'

import {
    Button,
    Checkbox,
    ButtonGroup,
    Position,
    useAccordion,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Icon,
    Typography,
} from '@sourcegraph/wildcard'

import { ExecutionOptions } from '../BatchSpecContext'

import styles from './RunBatchSpecButton.module.scss'

interface RunBatchSpecButtonProps {
    execute: () => void
    /**
     * Whether or not the button should be disabled. An optional tooltip string to display
     * may be provided in place of `true`.
     */
    isExecutionDisabled?: boolean | string
    options: ExecutionOptions
    onChangeOptions: (newOptions: ExecutionOptions) => void
}

export const RunBatchSpecButton: React.FunctionComponent<React.PropsWithChildren<RunBatchSpecButtonProps>> = ({
    execute,
    isExecutionDisabled = false,
    options,
    onChangeOptions,
}) => {
    const [isOpen, setIsOpen] = useState(false)

    return (
        // We need to use `Popover` instead of `Menu` because `MenuList` doesn't support
        // native tab indexing through the children; the parent holds focus instead,
        // similarly to a native dropdown selector.
        <Popover isOpen={isOpen} onOpenChange={event => setIsOpen(event.isOpen)}>
            <ButtonGroup className="mb-2">
                <Button
                    variant="primary"
                    onClick={execute}
                    disabled={!!isExecutionDisabled}
                    data-tooltip={typeof isExecutionDisabled === 'string' ? isExecutionDisabled : undefined}
                >
                    Run batch spec
                </Button>
                <PopoverTrigger
                    as={Button}
                    variant="primary"
                    type="button"
                    className={styles.executionOptionsMenuButton}
                >
                    <ChevronDownIcon />
                    <VisuallyHidden>Options</VisuallyHidden>
                </PopoverTrigger>
            </ButtonGroup>

            <PopoverContent className={styles.menuList} position={Position.bottomEnd}>
                <Typography.H3 className="pb-2 pt-3 pl-3 pr-3 m-0">Execution options</Typography.H3>
                <ExecutionOption moreInfo="When this batch spec is executed, it will not use cached results from any previous execution.">
                    <Checkbox
                        name="run-without-cache"
                        id="run-without-cache"
                        checked={options.runWithoutCache}
                        onChange={() => onChangeOptions({ runWithoutCache: !options.runWithoutCache })}
                        label="Run without cache"
                    />
                </ExecutionOption>
                <ExecutionOption disabled={true} disabledTooltip="Coming soon">
                    <Checkbox
                        name="apply-when-complete"
                        id="apply-when-complete"
                        checked={false}
                        disabled={true}
                        label="Apply when complete"
                    />
                </ExecutionOption>
            </PopoverContent>
        </Popover>
    )
}

type ExecutionOptionProps =
    | {
          disabled?: false
          moreInfo?: string
      }
    | {
          disabled: true
          disabledTooltip: string
      }

const ExecutionOption: React.FunctionComponent<React.PropsWithChildren<ExecutionOptionProps>> = props => {
    const [infoReference, infoOpen, setInfoOpen, infoStyle] = useAccordion<HTMLParagraphElement>()

    const info = props.disabled ? (
        <Icon className="ml-2" data-tooltip={props.disabledTooltip} tabIndex={0} as={InfoCircleOutlineIcon} />
    ) : props.moreInfo ? (
        <Button className="m-0 ml-2 p-0 border-0" onClick={() => setInfoOpen(!infoOpen)}>
            <Icon aria-hidden={true} as={InfoCircleOutlineIcon} />
            <VisuallyHidden>More info</VisuallyHidden>
        </Button>
    ) : null

    return (
        <div className={styles.executionOption}>
            <div className={styles.executionOptionForm}>
                {props.children}
                {info}
            </div>
            {!props.disabled && props.moreInfo && (
                <animated.div className={styles.expandedInfo} style={infoStyle}>
                    <p className="m-0 pb-2" ref={infoReference}>
                        {props.moreInfo}
                    </p>
                </animated.div>
            )}
        </div>
    )
}
