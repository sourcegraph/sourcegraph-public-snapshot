import { createContext } from 'react'

interface LineChartSettingsContext {
    zeroYAxisMin: boolean
    layout: 'vertical' | 'horizontal'
}

export const LineChartSettingsContext = createContext<LineChartSettingsContext>({
    zeroYAxisMin: false,
    layout: 'vertical',
})
