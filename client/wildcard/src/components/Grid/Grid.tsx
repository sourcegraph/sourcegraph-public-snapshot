import React, { CSSProperties } from 'react'

export interface GridProps {
    className?: string
    /**
     * The number of grid columns to render.
     *
     * @default 3
     */
    columnCount?: number
    /**
     * Rem Spacing between grid columns.
     *
     * @default 1
     */
    spacing?: number
}

/**
 * Dynamically generate the <Grid> style from provided props
 */
const getGridStyle = (columnCount: number, spacing: number): CSSProperties => ({
    display: 'grid',
    gap: `${spacing}rem`,
    gridTemplateColumns: `[row-start] repeat(${columnCount}, 1fr) [row-end]`,
    marginBottom: `${spacing}rem`,
})

/** A simple Grid component. Can be configured to display a number of columns with different gutter spacing. */
export const Grid: React.FunctionComponent<React.PropsWithChildren<GridProps>> = ({
    children,
    columnCount = 3,
    spacing = 1,
    className,
}) => (
    <div
        // We use `style` here to dynamically generate the grid styles.
        // eslint-disable-next-line react/forbid-dom-props
        style={getGridStyle(columnCount, spacing)}
        className={className}
    >
        {children}
    </div>
)
