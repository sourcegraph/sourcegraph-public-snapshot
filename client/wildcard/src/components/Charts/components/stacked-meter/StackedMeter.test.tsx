import { describe, expect, test } from 'vitest'

import { renderWithBrandedContext } from '../../../../testing'

import { StackedMeter } from './StackedMeter'

describe('<StackedMeter />', () => {
    test('renders correctly', () => {
        const { asFragment } = renderWithBrandedContext(
            <StackedMeter
                width={100}
                height={10}
                viewMinMax={[0, 100]}
                data={[
                    { name: 'foo', color: 'red', value: 20 },
                    { name: 'bar', color: 'blue', value: 30 },
                ]}
                getDatumValue={datum => datum.value}
                getDatumName={datum => datum.name}
                getDatumColor={datum => datum.color}
                minBarWidth={10}
                className="myclass"
                rightToLeft={true}
            />
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
