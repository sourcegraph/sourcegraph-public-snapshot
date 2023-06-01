import { FC } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { BarChart, Card, H2, Text } from '@sourcegraph/wildcard'

import { AnalyticsDateRange, GetInstanceOwnStatsResult } from '../../../graphql-operations'
import { AnalyticsPageTitle } from '../../../site-admin/analytics/components/AnalyticsPageTitle'
import { ChartContainer } from '../../../site-admin/analytics/components/ChartContainer'
import {
    TimeSavedCalculator,
    TimeSavedCalculatorProps,
} from '../../../site-admin/analytics/components/TimeSavedCalculatorGroup'
import { ValueLegendList, ValueLegendListProps } from '../../../site-admin/analytics/components/ValueLegendList'

import { GET_INSTANCE_OWN_STATS } from './query'

interface OwnUsageDatum {
    ownershipReasonType: string
    entriesCount: number
    fill: string
}

export const OwnAnalyticsPage: FC = () => {
    // const [localData, setLocalData] = useState<OwnSignalConfig[]>([])
    // const [saveError] = useState<Error | null>()

    const { data } = useQuery<GetInstanceOwnStatsResult>(GET_INSTANCE_OWN_STATS, {})

    const ownSignalsData: OwnUsageDatum[] = [
        {
            ownershipReasonType: 'Codeowned files',
            entriesCount: data?.instanceOwnershipStats?.totalCodeownedFiles || 0,
            fill: 'var(--info)',
        },
        {
            ownershipReasonType: 'Total files',
            entriesCount: data?.instanceOwnershipStats?.totalFiles || 0,
            fill: 'var(--text-muted)',
        },
    ]
    const getValue = (datum: OwnUsageDatum) => datum.entriesCount
    const getColor = (datum: OwnUsageDatum) => datum.fill
    const getLink = (datum: OwnUsageDatum) => ''
    const getName = (datum: OwnUsageDatum) => datum.ownershipReasonType
    const getGroup = (datum: OwnUsageDatum) => ''

    const legends: ValueLegendListProps['items'] = [
        {
            value: 48,
            description: 'Recent contributions',
            color: 'var(--cyan)',
            tooltip: 'Number of commits to this codebase',
        },
        {
            value: 720,
            description: 'Recent views',
            color: 'var(--orange)',
            tooltip: 'Total recent views',
        },
        {
            value: 5,
            description: 'Assigned owners',
            color: 'var(--merged)',
            position: 'right',
            tooltip: 'Number of owners assigned through the UI',
        },
    ]

    const calculatorProps: TimeSavedCalculatorProps = {
        page: 'Notebooks',
        label: 'Views',
        color: 'var(--body-color)',
        dateRange: AnalyticsDateRange.LAST_THREE_MONTHS,
        value: 774,
        defaultMinPerItem: 5,
        description:
            'Notebooks reduce the time it takes to create living documentation and share it. Each notebook view accounts for time saved by both creators and consumers of notebooks.',
        temporarySettingsKey: 'search.notebooks.minSavedPerView',
    }

    // const [saveConfigs, { loading: loadingSaveConfigs }] = useMutation<
    //     UpdateSignalConfigsResult,
    //     UpdateSignalConfigsVariables
    // >(UPDATE_SIGNAL_CONFIGURATIONS, {})

    // function onUpdateJob(index: number, newJob: OwnSignalConfig): void {
    //     setHasLocalChanges(true)
    //     const newData = localData.map((job: OwnSignalConfig, ind: number) => {
    //         if (ind === index) {
    //             return newJob
    //         }
    //         return job
    //     })
    //     setLocalData(newData)
    // }

    return (
        <>
            <AnalyticsPageTitle>Own</AnalyticsPageTitle>

            <Card className="p-3 position-relative">
                {/* <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                </div> */}
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {ownSignalsData && (
                    <div>
                        <ChartContainer
                            title={
                                'Title'
                                // aggregation.selected === 'count'
                                //     ? `${groupingLabel} activity`
                                //     : `${groupingLabel} unique users`
                            }
                            labelX="Time"
                            labelY={
                                'LabelY'
                                // aggregation.selected === 'count' ? 'Activity' : 'Unique users'
                            }
                        >
                            {width => (
                                <BarChart
                                    width={width}
                                    height={300}
                                    data={ownSignalsData}
                                    getDatumName={getName}
                                    getDatumValue={getValue}
                                    getDatumColor={getColor}
                                    getDatumLink={getLink}
                                    getDatumHover={datum => `custom text for ${datum.ownershipReasonType}`}
                                />
                            )}
                        </ChartContainer>
                        {/*
                        This is grouping toggle - maybe later?
                        <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                            <HorizontalSelect<typeof grouping.value> {...grouping} className="mr-4" />
                            <ToggleSelect<typeof aggregation.selected> {...aggregation} />
                        </div> */}
                    </div>
                )}
                <H2 className="my-3">Total time saved</H2>
                {calculatorProps && <TimeSavedCalculator {...calculatorProps} />}
                {/* <div className={styles.suggestionBox}>
                    <H4 className="my-3">Suggestions</H4>
                    <div className={classNames(styles.border, 'mb-3')} />
                    <ul className="mb-3 pl-3">
                        <Text as="li">
                            <AnchorLink to="https://about.sourcegraph.com/blog/notebooks-ci" target="_blank">
                                Learn more
                            </AnchorLink>{' '}
                            about how notebooks improves onbaording, code reuse and saves developers time.
                        </Text>
                    </ul>
                </div> */}
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )

    // return (
    //     <div>
    //         <span className={styles.topHeader}>
    //             <div>
    //                 <PageTitle title="Own Analytics" />
    //                 <PageHeader
    //                     headingElement="h2"
    //                     path={[{ text: 'Own Analytics' }]}
    //                     description="TODO"
    //                     className="mb-3"
    //                 />
    //                 {saveError && <ErrorAlert error={saveError} />}
    //             </div>
    //         </span>

    //         <Container className={styles.root}>
    //             {loading && <LoadingSpinner />}
    //             {error && <ErrorAlert prefix="Error fetching Own signal configurations" error={error} />}
    //             {!loading && localData && !error && (
    //                 <BarChart
    //                     width={400}
    //                     height={400}
    //                     data={ownSignalsData}
    //                     getDatumName={getName}
    //                     getDatumValue={getValue}
    //                     getDatumColor={getColor}
    //                     getDatumLink={getLink}
    //                     getDatumHover={datum => `custom text for ${datum.ownershipReasonType}`}
    //                 />
    //             )}
    //         </Container>
    //     </div>
    // )
}
