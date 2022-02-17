import VisuallyHidden from '@reach/visually-hidden'
import classNames from 'classnames'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React from 'react'
import { animated } from 'react-spring'

import {
    Button,
    Menu,
    MenuButton,
    Checkbox,
    ButtonGroup,
    MenuList,
    Position,
    useAccordion,
} from '@sourcegraph/wildcard'

import styles from './ExecutionOptions.module.scss'

export interface ExecutionOptions {
    runWithoutCache: boolean
}

interface ExecutionOptionsDropdownProps {
    execute: () => void
    isExecutionDisabled: boolean
    executionTooltip?: string
    options: ExecutionOptions
    onChangeOptions: (newOptions: ExecutionOptions) => void
}

export const ExecutionOptionsDropdown: React.FunctionComponent<ExecutionOptionsDropdownProps> = ({
    execute,
    isExecutionDisabled,
    executionTooltip,
    options,
    onChangeOptions,
}) => (
    <Menu>
        <ButtonGroup className="mb-2">
            <Button variant="primary" onClick={execute} disabled={isExecutionDisabled} data-tooltip={executionTooltip}>
                Run batch spec
            </Button>
            <MenuButton variant="primary" className={classNames(styles.executionOptionsMenuButton, 'dropdown-toggle')}>
                <VisuallyHidden>Options</VisuallyHidden>
            </MenuButton>
        </ButtonGroup>
        <MenuList className={styles.menuList} position={Position.bottomEnd}>
            <h3 className="pb-2 pt-3 pl-3 pr-3 m-0">Execution options</h3>
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
        </MenuList>
    </Menu>
)

interface ExecutionOptionProps {
    disabled?: boolean
    disabledTooltip?: string
    moreInfo?: string
}

const ExecutionOption: React.FunctionComponent<ExecutionOptionProps> = ({
    disabled = false,
    disabledTooltip,
    moreInfo,
    children,
}) => {
    const [infoReference, infoOpen, setInfoOpen, infoStyle] = useAccordion<HTMLParagraphElement>()

    const info =
        disabled && !!disabledTooltip ? (
            <InfoCircleOutlineIcon className="icon-inline ml-2" data-tooltip="Coming soon" />
        ) : moreInfo ? (
            <Button className="m-0 ml-2 p-0 border-0" onClick={() => setInfoOpen(!infoOpen)}>
                <InfoCircleOutlineIcon className="icon-inline" aria-hidden={true} />
                <VisuallyHidden>More info</VisuallyHidden>
            </Button>
        ) : null

    return (
        <div className={styles.executionOption}>
            <div className={styles.executionOptionForm}>
                {children}
                {info}
            </div>
            {!disabled && moreInfo && (
                <animated.div className={styles.expandedInfo} style={infoStyle}>
                    <p className="m-0 pb-2" ref={infoReference}>
                        {moreInfo}
                    </p>
                </animated.div>
            )}
        </div>
    )
}
