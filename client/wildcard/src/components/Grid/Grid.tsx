import React, { type CSSProperties } from 'react'

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
    spacing?: number | number[]
    /**
     * optional CSS definition of grid-template-columns property
     */
    templateColumns?: string
}

/**
 * Dynamically generate the <Grid> style from provided props
 */
const getGridStyle = (
    columnCount: GridProps['columnCount'],
    spacing: GridProps['spacing'],
    templateColumns?: GridProps['templateColumns']
): CSSProperties => ({
    display: 'grid',
    gap: Array.isArray(spacing) && spacing.length === 2 ? `${spacing[0]}rem ${spacing[1]}rem` : `${spacing}rem`,
    gridTemplateColumns: templateColumns ?? `[row-start] repeat(${columnCount}, 1fr) [row-end]`,
    marginBottom: `${spacing}rem`,
})

/** A simple Grid component. Can be configured to display a number of columns with different gutter spacing. */
export const Grid: React.FunctionComponent<React.PropsWithChildren<GridProps>> = ({
    children,
    columnCount = 3,
    spacing = 1,
    className,
    templateColumns,
}) => (
    <div
        // We use `style` here to dynamically generate the grid styles.
        // eslint-disable-next-line react/forbid-dom-props
        style={getGridStyle(columnCount, spacing, templateColumns)}
        className={className}
    >
        {children}
    </div>
)
