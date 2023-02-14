import React, { useState } from 'react'

import classNames from 'classnames'

import { Card, CardHeader, H3, Popover, PopoverContent, Position } from '@sourcegraph/wildcard'

import { CodyIcon } from '../../CodyIcon'

import styles from './CodyAssistancePopover.module.scss'

interface Props {
    openByDefault?: boolean
    position?: Position
    triggerElement: React.ReactNode
    content: React.ReactNode
    popoverClassName?: string
}

export const CodyAssistancePopover: React.FunctionComponent<Props> = ({
    openByDefault,
    position = Position.bottomEnd,
    triggerElement,
    content,
    popoverClassName,
}) => {
    const [isOpen, setIsOpen] = useState(!!openByDefault)

    return (
        <Popover isOpen={isOpen} onOpenChange={event => setIsOpen(event.isOpen)}>
            {triggerElement}
            <PopoverContent position={position} className={classNames(styles.popoverContent, popoverClassName)}>
                <Card className={styles.card}>
                    <CardHeader className={styles.cardHeader}>
                        <H3 className="mb-0">
                            <CodyIcon /> Cody
                        </H3>
                    </CardHeader>
                    {content}
                </Card>
            </PopoverContent>
        </Popover>
    )
}
