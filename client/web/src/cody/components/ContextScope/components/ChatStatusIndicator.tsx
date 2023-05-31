import { useState } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import {
    Button,
    Icon,
    Popover,
    PopoverContent,
    PopoverTail,
    PopoverTrigger,
    Position,
    Tooltip,
} from '@sourcegraph/wildcard'

import styles from './ChatStatusIndicator.module.scss'

interface ChatStatusIndicatorProps {
    status: 'low' | 'medium' | 'high'
}

export const ChatStatusIndicator: React.FC<ChatStatusIndicatorProps> = ({ status = 'low' }) => {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)

    const messages = [
        {
            status: 'low',
            content: (
                <>
                    <span style={{ fontWeight: 600 }}>
                        Cody indexing has not been enabled for the repositories you selected.
                    </span>{' '}
                    To enable it, please contact your admin. You can still use Cody, but the quality of Cody’s responses
                    may be low.
                </>
            ),
        },
        {
            status: 'medium',
            content: (
                <>
                    <span style={{ fontWeight: 600 }}>
                        Repositories are still being indexed and will be ready soon.
                    </span>{' '}
                    You can still use Cody while you wait, but the quality of Cody’s responses may be low.
                </>
            ),
        },
        {
            status: 'high',
            content: (
                <>
                    <span style={{ fontWeight: 600 }}>All selected repositories are indexed.</span> When embeddings are
                    running, you’ll see how amazing the results are!
                </>
            ),
        },
    ]

    const selectedMessage = messages.find(message => message.status === status)?.content
    const title = `Cody's response quality is ${status}`

    return (
        <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
            <Tooltip content={title}>
                <PopoverTrigger as={Button} outline={false} className="button">
                    <CodyStatusIcon status={status} />
                </PopoverTrigger>
            </Tooltip>

            <PopoverContent position={Position.topStart} className={styles.wrapper}>
                <div className={classNames('justify-content-between', styles.header)}>
                    {title}{' '}
                    <Button onClick={() => setIsPopoverOpen(false)} variant="icon" aria-label="Close">
                        <Icon aria-hidden={true} svgPath={mdiClose} />
                    </Button>
                </div>
                <div className={styles.contentBody}>
                    <div className={styles.leftPanel}>Embeddings:</div>
                    <div className={styles.rightPanel}>{selectedMessage}</div>
                </div>
            </PopoverContent>

            <PopoverTail size="sm" />
        </Popover>
    )
}

interface CodyStatusIconProps {
    status: 'low' | 'medium' | 'high'
}

const CodyStatusIcon: React.FC<CodyStatusIconProps> = ({ status }) => {
    if (status === 'low') {
        return (
            <svg width="15" height="16" viewBox="0 0 19 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                    d="M16 10.09V4C16 1.79 12.42 0 8 0C3.58 0 0 1.79 0 4V14C0 16.21 3.59 18 8 18C8.46 18 8.9 18 9.33 17.94C9.12 17.33 9 16.68 9 16V15.95C8.68 16 8.35 16 8 16C4.13 16 2 14.5 2 14V11.77C3.61 12.55 5.72 13 8 13C8.65 13 9.27 12.96 9.88 12.89C10.93 11.16 12.83 10 15 10C15.34 10 15.67 10.04 16 10.09ZM14 9.45C12.7 10.4 10.42 11 8 11C5.58 11 3.3 10.4 2 9.45V6.64C3.47 7.47 5.61 8 8 8C10.39 8 12.53 7.47 14 6.64V9.45ZM8 6C4.13 6 2 4.5 2 4C2 3.5 4.13 2 8 2C11.87 2 14 3.5 14 4C14 4.5 11.87 6 8 6ZM16.41 16L18.54 18.12L17.12 19.54L15 17.41L12.88 19.54L11.47 18.12L13.59 16L11.47 13.88L12.88 12.47L15 14.59L17.12 12.47L18.54 13.88L16.41 16Z"
                    fill="#E67700"
                />
            </svg>
        )
    } else if (status === 'medium') {
        return (
            <svg width="15" height="16" viewBox="0 0 19 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                    d="M8 13C8.41 13 8.81 12.97 9.21 12.94C9.4 12.18 9.72 11.46 10.16 10.83C9.47 10.94 8.74 11 8 11C5.58 11 3.3 10.4 2 9.45V6.64C3.47 7.47 5.61 8 8 8C10.39 8 12.53 7.47 14 6.64V8.19C14.5 8.07 15 8 15.55 8C15.7 8 15.85 8 16 8.03V4C16 1.79 12.42 0 8 0C3.58 0 0 1.79 0 4V14C0 16.21 3.59 18 8 18C8.66 18 9.31 17.96 9.92 17.88C9.57 17.29 9.31 16.64 9.16 15.94C8.79 16 8.41 16 8 16C4.13 16 2 14.5 2 14V11.77C3.61 12.55 5.72 13 8 13ZM8 2C11.87 2 14 3.5 14 4C14 4.5 11.87 6 8 6C4.13 6 2 4.5 2 4C2 3.5 4.13 2 8 2ZM19 14.5C19 15.32 18.75 16.08 18.33 16.71L17.24 15.62C17.41 15.28 17.5 14.9 17.5 14.5C17.5 13.12 16.38 12 15 12V13.5L12.75 11.25L15 9V10.5C17.21 10.5 19 12.29 19 14.5ZM15 15.5L17.25 17.75L15 20V18.5C12.79 18.5 11 16.71 11 14.5C11 13.68 11.25 12.92 11.67 12.29L12.76 13.38C12.59 13.72 12.5 14.1 12.5 14.5C12.5 15.88 13.62 17 15 17V15.5Z"
                    fill="#798BAF"
                />
            </svg>
        )
    } else if (status === 'high') {
        return (
            <svg width="15" height="15" viewBox="0 0 19 19" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                    d="M16 10.09V4C16 1.79 12.42 0 8 0C3.58 0 0 1.79 0 4V14C0 16.21 3.59 18 8 18C8.46 18 8.9 18 9.33 17.94C9.1129 17.3162 9.00137 16.6605 9 16V15.95C8.68 16 8.35 16 8 16C4.13 16 2 14.5 2 14V11.77C3.61 12.55 5.72 13 8 13C8.65 13 9.27 12.96 9.88 12.89C10.4127 12.0085 11.1638 11.2794 12.0607 10.7731C12.9577 10.2668 13.9701 10.0005 15 10C15.34 10 15.67 10.04 16 10.09ZM14 9.45C12.7 10.4 10.42 11 8 11C5.58 11 3.3 10.4 2 9.45V6.64C3.47 7.47 5.61 8 8 8C10.39 8 12.53 7.47 14 6.64V9.45ZM8 6C4.13 6 2 4.5 2 4C2 3.5 4.13 2 8 2C11.87 2 14 3.5 14 4C14 4.5 11.87 6 8 6ZM18.5 14.25L13.75 19L11 16L12.16 14.84L13.75 16.43L17.34 12.84L18.5 14.25Z"
                    fill="url(#paint0_linear_48_6629)"
                />
                <defs>
                    <linearGradient
                        id="paint0_linear_48_6629"
                        x1="9.25"
                        y1="0"
                        x2="9.25"
                        y2="19"
                        gradientUnits="userSpaceOnUse"
                    >
                        <stop stop-color="#A112FF" />
                        <stop offset="0.463542" stop-color="#FF5543" />
                        <stop offset="1" stop-color="#00CBEC" />
                    </linearGradient>
                </defs>
            </svg>
        )
    }
}
