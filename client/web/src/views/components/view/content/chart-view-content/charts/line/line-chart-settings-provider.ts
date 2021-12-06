import { createContext } from 'react'

export enum LineChartLayoutOrientation {
    Vertical = 'vertical',
    Horizontal = 'horizontal',
}

interface LineChartSettingsContext {
    /**
     * Enables Y-label generation from 0 value.
     *
     * ```
     * With zeroYAxisMin: false       With zeroYAxisMin: true
     *   ▲                            ▲
     * 40│             ● ●          45│
     *   │      ●     ●               │           ● ● ●
     * 30│     ●  ●  ●              30│    ● ●   ●
     *   │    ●    ●                  │   ●    ●
     * 20│   ●                      15│  ●
     *   │                            │
     * 10└─────────────────▶        0 └─────────────────▶
     *```
     */
    zeroYAxisMin: boolean

    /**
     * Controls chart legend layout position. If it's property isn't specified
     * line chart uses its internal logic about putting legend block which is based
     * on chart width and number of lines (series).
     *
     * ```
     * Vertical (default)      Horizontal
     * ▲                       ▲
     * │             ● ●       │           ●    ● Item 1
     * │      ●     ●          │    ●     ●     ● Item 2
     * │     ●  ●  ●           │   ●  ●  ●
     * │    ●    ●             │  ●    ●
     * │   ●                   │ ●
     * │                       │
     * └─────────────────▶     └─────────────▶
     * ● Item 1 ● Item 2
     * ```
     */
    layout?: LineChartLayoutOrientation
}

export const LineChartSettingsContext = createContext<LineChartSettingsContext>({
    zeroYAxisMin: false,
})
