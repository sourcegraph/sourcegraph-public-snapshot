/**
 * Default value for line color in case if we didn't get color for line from content config.
 */
export const DEFAULT_LINE_STROKE = 'var(--gray-07)'

/**
 * Visx xy-chart supports data series with missing. To show the
 * points but not the very beginning of the chart we should use
 * this default empty value. See example below points that have
 * EMPTY_DATA_POINT_VALUE value haven't been rendered instead of
 * that we rendered non active background (``` area)
 *
 * <pre>
 * ┌──────────────────────────────┐     ┌──────────────────────────────┐
 * │``````````````                │ 10  │         ````````             │ 10
 * │``````````````                │     │       ▼ ````````             │
 * │``````````````              ▼ │ 9   │         ````````           ▼ │ 9
 * │``````````````                │     │  ▼      ````````             │
 * │``````````````      ▼         │ 8   │         ````````    ▼        │ 8
 * │``````````````                │     │         ````````             │
 * │``````````````          ▼     │ 7   │         ````````       ▼     │ 7
 * │`````````````` ▼              │     │    ▼    ```````` ▼           │
 * │``````````````                │ 6   │         ````````             │ 6
 * │``````````````                │     │         ````````             │
 * │``````````````                │ 5   │         ````````             │ 5
 * └──────────────────────────────┘     └──────────────────────────────┘
 * </pre>
 */
export const EMPTY_DATA_POINT_VALUE = null

/**
 * If width of the chart is less than this var width value we should put the legend
 * block below the chart block
 *
 * ```
 * Less than 450px - put legend below      Chart block has enough space - render legend aside
 * ▲                                       ▲
 * │             ● ●                       │           ●    ● Item 1
 * │      ●     ●                          │    ●     ●     ● Item 2
 * │     ●  ●  ●                           │   ●  ●  ●
 * │    ●    ●                             │  ●    ●
 * │   ●                                   │ ●
 * │                                       │
 * └─────────────────▶                     └─────────────▶
 * ● Item 1 ● Item 2
 * ```
 */
export const MINIMAL_HORIZONTAL_LAYOUT_WIDTH = 460

/**
 * Even if have a big enough width for putting legend aside (see {@link MINIMAL_HORIZONTAL_LAYOUT_WIDTH})
 * we should enable this mode only if line chart has more than 3 series
 */
export const MINIMAL_SERIES_FOR_ASIDE_LEGEND = 3
