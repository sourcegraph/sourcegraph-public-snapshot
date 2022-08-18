import { FC, useCallback, useMemo } from 'react'

import { mdiArrowExpand, mdiPlus } from '@mdi/js'
import { ParentSize } from '@visx/responsive'
import { getTicks } from '@visx/scale'
import { AnyD3Scale } from '@visx/scale/lib/types/Scale'
import { useHistory, useLocation } from 'react-router'

import { ButtonGroup, Button, Icon, BarChart } from '@sourcegraph/wildcard'

import { LANGUAGE_USAGE_DATA, LanguageUsageDatum } from './search-aggregation-mock-data'

import styles from './SearchAggregations.module.scss'

interface URLStateOptions<State> {
    urlKey: string
    deserializer: (value: string | null) => State
    serializer: (state: State) => string
}

type SetStateResult<State> = [state: State, dispatch: (state: State) => void]

/**
 * React hook analog standard react useState hook but with synced value with URL
 * through URL query parameter.
 */
function useSyncedWithURLState<State>(options: URLStateOptions<State>): SetStateResult<State> {
    const { urlKey, serializer, deserializer } = options
    const history = useHistory()
    const { search } = useLocation()

    const urlSearchParameters = useMemo(() => new URLSearchParams(search), [search])
    const queryParameter = useMemo(() => deserializer(urlSearchParameters.get(urlKey)), [
        urlSearchParameters,
        urlKey,
        deserializer,
    ])

    const setNextState = useCallback(
        (nextState: State) => {
            urlSearchParameters.set(urlKey, serializer(nextState))

            history.replace({ search: `?${urlSearchParameters.toString()}` })
        },
        [history, serializer, urlKey, urlSearchParameters]
    )

    return [queryParameter, setNextState]
}

enum AggregationModes {
    Repository = 'repo',
    FilePath = 'file',
    Author = 'author',
    CaptureGroups = 'captureGroup',
}

const AGGREGATION_MODE_URL_KEY = 'groupBy'

const aggregationModeDeserializer = (serializedValue: string | null): AggregationModes => {
    switch (serializedValue) {
        case 'repo':
            return AggregationModes.Repository
        case 'file':
            return AggregationModes.FilePath
        case 'author':
            return AggregationModes.Author
        case 'captureGroup':
            return AggregationModes.CaptureGroups

        default:
            return AggregationModes.Repository
    }
}

const aggregationModeSerializer = (mode: AggregationModes): string => mode.toString()

const getValue = (datum: LanguageUsageDatum): number => datum.value
const getColor = (datum: LanguageUsageDatum): string => datum.fill
const getLink = (datum: LanguageUsageDatum): string => datum.linkURL
const getName = (datum: LanguageUsageDatum): string => datum.name

interface GetScaleTicksOptions {
    scale: AnyD3Scale
    space: number
    pixelsPerTick?: number
}

/**
 * Custom tick generator for search result aggregation bar. Double down tick
 * labels count every until their count fits in a given available space.
 *
 * ```
 * Before
 * ─┬──┬──┬──┬──┬──┬──┬──┬──┬──┬───▶
 *  1  2  3  4  5  6  7  8  9  10
 *
 * After
 * ─┬─────┬─────┬─────┬─────┬──────▶
 *  1     3     5     7     9
 * ```
 */
const getXScaleTicks = <T,>(options: GetScaleTicksOptions): T[] => {
    const { scale, space, pixelsPerTick = 80 } = options

    // Calculate desirable number of ticks
    const numberTicks = Math.max(2, Math.floor(space / pixelsPerTick))

    let filteredTicks = getTicks(scale)

    while (filteredTicks.length > numberTicks) {
        filteredTicks = filteredTicks.filter((tick, index) => index % 2 === 0)
    }

    return filteredTicks
}

const MAX_TRUNCATED_LABEL_LENGTH = 10
const getTruncatedTick = (tick: string): string =>
    tick.length >= MAX_TRUNCATED_LABEL_LENGTH ? `${tick.slice(0, MAX_TRUNCATED_LABEL_LENGTH)}...` : tick
const getTruncatedTickFromTheEnd = (tick: string): string =>
    tick.length >= MAX_TRUNCATED_LABEL_LENGTH ? `...${tick.slice(-MAX_TRUNCATED_LABEL_LENGTH)}` : tick

const getTruncationFormatter = (aggregationMode: AggregationModes): ((tick: string) => string) => {
    switch (aggregationMode) {
        // These types possibly have long labels with the same pattern at the start of the string,
        // so we truncate their labels from the end
        case AggregationModes.Repository:
        case AggregationModes.FilePath:
            return getTruncatedTickFromTheEnd

        default:
            return getTruncatedTick
    }
}

interface SearchAggregationsProps {}

export const SearchAggregations: FC<SearchAggregationsProps> = props => {
    const [aggregationMode, setAggregationMode] = useSyncedWithURLState({
        urlKey: AGGREGATION_MODE_URL_KEY,
        serializer: aggregationModeSerializer,
        deserializer: aggregationModeDeserializer,
    })

    const getTruncatedXLabel = useMemo(() => getTruncationFormatter(aggregationMode), [aggregationMode])

    return (
        <article className="pt-2">
            <ButtonGroup className="mb-3">
                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.Repository}
                    onClick={() => setAggregationMode(AggregationModes.Repository)}
                >
                    Repo
                </Button>

                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.FilePath}
                    onClick={() => setAggregationMode(AggregationModes.FilePath)}
                >
                    File
                </Button>

                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.Author}
                    onClick={() => setAggregationMode(AggregationModes.Author)}
                >
                    Author
                </Button>
                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.CaptureGroups}
                    onClick={() => setAggregationMode(AggregationModes.CaptureGroups)}
                >
                    Capture group
                </Button>
            </ButtonGroup>

            <ParentSize className={styles.chartContainer}>
                {parent => (
                    <BarChart
                        width={parent.width}
                        height={parent.height}
                        data={LANGUAGE_USAGE_DATA}
                        getDatumName={getName}
                        getDatumValue={getValue}
                        getDatumColor={getColor}
                        getDatumLink={getLink}
                        pixelsPerYTick={20}
                        pixelsPerXTick={20}
                        maxAngleXTick={45}
                        getScaleXTicks={getXScaleTicks}
                        getTruncatedXTick={getTruncatedXLabel}
                    />
                )}
            </ParentSize>

            <footer className={styles.actions}>
                <Button variant="secondary" size="sm" outline={true} className={styles.detailsAction}>
                    <Icon aria-hidden={true} svgPath={mdiArrowExpand} /> Expand
                </Button>

                <Button variant="secondary" outline={true} size="sm">
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Save insight
                </Button>
            </footer>
        </article>
    )
}
