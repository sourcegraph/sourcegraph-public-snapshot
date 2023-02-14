import React from 'react'

import classNames from 'classnames'

import { PopoverTrigger } from '@sourcegraph/wildcard'

import { CodyAssistanceButton } from './CodyAssistanceButton'
import { CodyAssistancePopover } from './CodyAssistancePopover'

import styles from './GlobalCodyAssistancePopover.module.scss'

interface Props {}

export const GlobalCodyAssistancePopover: React.FunctionComponent<Props> = () => (
    <CodyAssistancePopover
        openByDefault={false}
        triggerElement={
            <PopoverTrigger
                as={CodyAssistanceButton}
                text={null}
                variant="purple"
                outline={true}
                className={classNames(styles.triggerButton, 'py-1 px-2')}
            />
        }
        content={<PopoverContent />}
        popoverClassName={styles.popover}
    />
)

const PopoverContent: React.FunctionComponent = () => (
    <ul className="list-group list-group-flush">
        <li className={classNames('list-group-item', styles.listGroupItem)}>foo</li>
        <li className={classNames('list-group-item', styles.listGroupItem)}>fofdfdfdo</li>
        <li className={classNames('list-group-item', styles.listGroupItem)}>foo334</li>
    </ul>
)
