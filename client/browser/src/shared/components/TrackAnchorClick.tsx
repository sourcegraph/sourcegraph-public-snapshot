import React, { useCallback } from 'react'

interface TrackAnchorClickProps {
    onClick: (event: React.MouseEvent) => void
    as?: keyof HTMLElementTagNameMap
}

/**
 * Track all anchor link clicks in children components
 */
export const TrackAnchorClick: React.FunctionComponent<React.PropsWithChildren<TrackAnchorClickProps>> = ({
    children,
    onClick,
    as: Tag = 'div',
}) => {
    const handleClick = useCallback(
        (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
            if ((event.target as HTMLElement)?.tagName.toLowerCase() === 'a') {
                onClick(event)
            }
        },
        [onClick]
    )
    return (
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        <Tag onClick={handleClick}>{children}</Tag>
    )
}
