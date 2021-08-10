import { createContext } from 'react'

interface LineChartSettingsContext {
    zeroYAxisMin: boolean
}

export const LineChartSettingsContext = createContext<LineChartSettingsContext>({ zeroYAxisMin: false })
