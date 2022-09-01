import { useCallback, useEffect } from 'react'

import classNames from 'classnames'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import { BehaviorSubject } from 'rxjs'

import {
    createRectangle,
    Icon,
    Link,
    Popover,
    PopoverContent,
    PopoverOpenEvent,
    PopoverTrigger,
    Position,
    useObservable,
} from '@sourcegraph/wildcard'

import { BlameHunk } from '../blame/useBlameHunks'

import styles from './BlameDecoration.module.scss'

const currentPopoverId = new BehaviorSubject<string | null>(null)
let timeoutId: NodeJS.Timeout | null = null
const resetTimeout = (): void => {
    if (timeoutId) {
        clearTimeout(timeoutId)
        timeoutId = null
    }
}

const usePopover = ({
    id,
    timeout,
    onOpen,
    onClose,
}: {
    id: string
    timeout: number
    onOpen?: () => void
    onClose?: () => void
}): {
    isOpen: boolean
    open: () => void
    close: () => void
    closeWithTimeout: () => void
    resetCloseTimeout: () => void
} => {
    const popoverId = useObservable(currentPopoverId)

    const isOpen = popoverId === id
    useEffect(() => {
        if (isOpen) {
            onOpen?.()
        }

        return () => {
            if (isOpen) {
                onClose?.()
            }
        }
    }, [isOpen, onOpen, onClose])

    const open = useCallback(() => {
        resetTimeout()
        currentPopoverId.next(id)
    }, [id])

    const close = useCallback(() => {
        if (currentPopoverId.getValue() === id) {
            currentPopoverId.next(null)
        }
    }, [id])

    const closeWithTimeout = useCallback(() => {
        timeoutId = setTimeout(close, timeout)
    }, [close, timeout])

    return { isOpen, open, close, closeWithTimeout, resetCloseTimeout: resetTimeout }
}

export const BlameDecoration: React.FunctionComponent<{
    line: number // 1-based line number
    blameHunk?: BlameHunk
    onSelect?: (line: number) => void
    onDeselect?: (line: number) => void
}> = ({ line, blameHunk, onSelect, onDeselect }) => {
    const id = line?.toString() || ''
    const onOpen = useCallback(() => onSelect?.(line), [onSelect, line])
    const onClose = useCallback(() => onDeselect?.(line), [onDeselect, line])
    const { isOpen, open, close, closeWithTimeout, resetCloseTimeout } = usePopover({
        id,
        timeout: 1000,
        onOpen,
        onClose,
    })

    const onPopoverOpenChange = useCallback((event: PopoverOpenEvent) => (event.isOpen ? close() : open()), [
        close,
        open,
    ])

    if (!blameHunk) {
        return null
    }

    return (
        <Popover isOpen={isOpen} onOpenChange={onPopoverOpenChange} key={id}>
            <PopoverTrigger
                as={Link}
                to={blameHunk.displayInfo.linkURL}
                target="_blank"
                rel="noreferrer noopener"
                className={classNames(styles.popoverTrigger, 'px-2')}
                onFocus={open}
                onBlur={close}
                onMouseEnter={open}
                onMouseLeave={closeWithTimeout}
            >
                <span
                    className={styles.content}
                    data-line-decoration-attachment-content={true}
                    data-contents={blameHunk.displayInfo.message}
                />
            </PopoverTrigger>

            <PopoverContent
                targetPadding={createRectangle(0, 0, 8, 8)}
                position={Position.topStart}
                focusLocked={false}
                onMouseEnter={resetCloseTimeout}
                onMouseLeave={close}
                className={styles.popoverContent}
            >
                <div className="py-1">
                    <div className="py-2 px-3">
                        <span className={styles.author}>{blameHunk.displayInfo.displayName}</span>{' '}
                        {blameHunk.displayInfo.dateString}
                    </div>
                    <hr className={styles.separator} />
                    <div className="py-2 px-3 d-flex align-items-center">
                        <Icon aria-hidden={true} as={SourceCommitIcon} className="mr-2 flex-shrink-0" />
                        <Link
                            to={blameHunk.displayInfo.linkURL}
                            target="_blank"
                            rel="noreferrer noopener"
                            className={styles.link}
                        >
                            {blameHunk.message}
                        </Link>
                    </div>
                </div>
            </PopoverContent>
        </Popover>
    )
}
