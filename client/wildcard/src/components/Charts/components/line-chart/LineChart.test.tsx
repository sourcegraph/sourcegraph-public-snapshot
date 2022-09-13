import { render } from '@testing-library/react'

import { LineChart } from './LineChart'

describe('LineChart', () => {
    it('should render', () => {
        render(<LineChart width={0} height={0} series={[]} />)
    })
})
