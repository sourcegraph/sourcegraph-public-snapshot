import { useMemo, useState } from 'react'

import { Meta, Story } from '@storybook/react'
import { AxisScale } from '@visx/axis/lib/types'
import { ParentSize } from '@visx/responsive'
import { scaleBand, scaleLinear, scaleTime } from '@visx/scale'

import { WebStory } from '../../../components/WebStory'

import { formatDateTick } from './axis/tick-formatters'
import { SvgRoot, SvgAxisLeft, SvgAxisBottom, SvgContent } from './SvgRoot'

const StoryConfig: Meta = {
    title: 'web/charts/axis',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default StoryConfig

const X_SCALE: AxisScale = scaleTime<number>({
    domain: [new Date(2022, 8, 22), new Date(2022, 10, 22)],
    nice: true,
    clamp: true,
})

const X_SCALE2: AxisScale = scaleBand<string>({
    domain: ['hello', 'worlddddddddddd', 'this', 'is', 'rotation', 'speaking'],
    padding: 0.2,
})

export const Demo: Story = props => {
    const [withBigYScale, switchYScale] = useState(true)

    const Y_SCALE: AxisScale = useMemo(
        () =>
            scaleLinear({
                domain: [0, withBigYScale ? 10000 : 1000000000000000000000000000000000000],
                nice: true,
                clamp: true,
            }),
        [withBigYScale]
    )
    return (
        <main>
            <button onClick={() => switchYScale(withBigYScale => !withBigYScale)}>Change dataset</button>
            <section style={{ display: 'flex', gap: 20, flexWrap: 'wrap' }}>
                <ParentSize style={{ width: 400, height: 400 }} debounceTime={0} className="flex-shrink-0">
                    {parent => (
                        <SvgRoot width={parent.width} height={parent.height} xScale={X_SCALE} yScale={Y_SCALE}>
                            <SvgAxisLeft />
                            <SvgAxisBottom<Date> tickFormat={formatDateTick} pixelsPerTick={20} />

                            <SvgContent>
                                {({ content }) => (
                                    <rect
                                        fill="darkslateblue"
                                        opacity={1}
                                        width={content.width}
                                        height={content.height}
                                    />
                                )}
                            </SvgContent>
                        </SvgRoot>
                    )}
                </ParentSize>

                <ParentSize style={{ width: 400, height: 400 }} debounceTime={0} className="flex-shrink-0">
                    {parent => (
                        <SvgRoot width={parent.width} height={parent.height} xScale={X_SCALE2} yScale={Y_SCALE}>
                            <SvgAxisLeft />
                            <SvgAxisBottom />

                            <SvgContent>
                                {({ content }) => <rect fill="pink" width={content.width} height={content.height} />}
                            </SvgContent>
                        </SvgRoot>
                    )}
                </ParentSize>

                <ParentSize style={{ width: 400, height: 400 }} debounceTime={0} className="flex-shrink-0">
                    {parent => (
                        <SvgRoot width={parent.width} height={parent.height} xScale={X_SCALE2} yScale={Y_SCALE}>
                            <SvgAxisLeft />
                            <SvgAxisBottom />

                            <SvgContent>
                                {({ content }) => (
                                    <rect fill="dodgerblue" width={content.width} height={content.height} />
                                )}
                            </SvgContent>
                        </SvgRoot>
                    )}
                </ParentSize>
            </section>
        </main>
    )
}
