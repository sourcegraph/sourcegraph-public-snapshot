import { createContext } from 'react'

interface LineChartSettingsContext {
    zeroYAxisMin: boolean
    toggleZeroYAxisMin?: () => void
}

export const LineChartSettingsContext = createContext<LineChartSettingsContext>({ zeroYAxisMin: false })
