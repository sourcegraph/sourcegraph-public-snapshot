import type { Meta, StoryFn } from '@storybook/react'

import { Grid, type GridProps } from './Grid'

const config: Meta = {
    title: 'wildcard/Grid',

    parameters: {
        component: Grid,
    },
}

export default config

export const GridExamples: StoryFn<GridProps> = ({ columnCount = 5, spacing }) => (
    <Grid columnCount={columnCount} spacing={spacing}>
        {/* Fill the grid with 20 items to showcase different setups */}
        {new Array(20).fill(0).map((_value, index) => (
            // eslint-disable-next-line react/no-array-index-key
            <div key={index}>Column</div>
        ))}
    </Grid>
)
