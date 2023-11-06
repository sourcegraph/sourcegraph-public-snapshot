import { describe, expect, it } from '@jest/globals'
import { render, screen, within } from '@testing-library/react'

import { LineChart } from './LineChart'
import { FLAT_SERIES } from './story/mocks'

const defaultArgs: RenderChartArgs = { series: FLAT_SERIES }

interface RenderChartArgs {
    series: typeof FLAT_SERIES
}

/**
 * Test padding set 1px to the left and bottom values in order to force
 * content sync appearance. In browser runtime this padding is calculated
 * based on chart axes sizes. In test environment size measurement API
 * doesn't work, we have to set padding manually in order to force chart
 * content appearance. See SVGContent component for more context.
 */
const TEST_PADDING = { top: 16, right: 18, bottom: 1, left: 1 }

const renderChart = ({ series }: RenderChartArgs) =>
    render(<LineChart width={400} height={400} series={series} padding={TEST_PADDING} />)

describe('LineChart', () => {
    // Non-exhaustive smoke tests to check that the chart renders correctly
    // All other general rendering tests are covered by chromatic
    describe('should render', () => {
        it('empty series', () => {
            renderChart({ ...defaultArgs, series: [] })
        })

        it('series with data', () => {
            renderChart(defaultArgs)

            // Query chart series list
            const series = screen.getByLabelText('Chart series')

            // Check that series were rendered
            const series1 = within(series).getByLabelText('A metric')
            const series2 = within(series).getByLabelText('C metric')
            const series3 = within(series).getByLabelText('B metric')
            expect(series1)
            expect(series2)
            expect(series3)

            // Check number of data points rendered
            expect(within(series1).getAllByRole('listitem')).toHaveLength(FLAT_SERIES[0].data.length)
            expect(within(series2).getAllByRole('listitem')).toHaveLength(FLAT_SERIES[1].data.length)
            expect(within(series3).getAllByRole('listitem')).toHaveLength(FLAT_SERIES[2].data.length)

            // Spot check y axis labels
            expect(screen.getByLabelText(/axis tick, value: 8/i)).toBeInTheDocument()
            expect(screen.getByLabelText(/axis tick, value: 20/i)).toBeInTheDocument()
            expect(screen.getByLabelText(/axis tick, value: 36/i)).toBeInTheDocument()

            // Spot check x axis labels
            expect(screen.getByLabelText(/axis tick, value: .*jan 01 2021/i)).toBeInTheDocument()
            expect(screen.getByLabelText(/axis tick, value: .*oct 01 2021/i)).toBeInTheDocument()
            expect(screen.getByLabelText(/axis tick, value: .*oct 01 2022/i)).toBeInTheDocument()
        })
    })

    describe('should handle clicks', () => {
        it('on a point', () => {
            renderChart(defaultArgs)

            // Query chart series list
            const series = screen.getByLabelText('Chart series')
            const [firstSeries] = within(series).getAllByRole('listitem')
            const [point00, point01, point02] = within(firstSeries).getAllByRole('listitem')

            // Spot checking multiple points
            // related issue https://github.com/sourcegraph/sourcegraph/issues/38304
            expect(point00).toHaveAttribute('href', 'https://yandex.com/search')
            expect(point01).toHaveAttribute('href', 'https://yandex.com/search')
            expect(point02).toHaveAttribute('href', 'https://yandex.com/search')
        })
    })
})
