import { Meta, Story } from '@storybook/react';
import { AxisScale } from '@visx/axis/lib/types';
import { ParentSize } from '@visx/responsive';
import { scaleBand, scaleLinear, scaleTime } from '@visx/scale';

import { WebStory } from '../../../components/WebStory';

import { SvgRoot, SvgAxisLeft, SvgAxisBottom, SvgContent } from './SvgRoot';
import { useMemo, useState } from 'react';

const StoryConfig: Meta = {
    title: 'web/charts/demo',
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
    const [on, setOn] = useState(true)
    const [rotate, setRotate]  = useState(true)

    const Y_SCALE: AxisScale = useMemo(() =>
            scaleLinear({
                domain: [0, on ? 10000 : 1000000000000000000000000000000000000],
                nice: true,
                clamp: true,
            }),
        [on]
    )
    return (
        <section style={{ display: 'flex', gap: 20 }}>
            <div>
                <button onClick={() => setOn(on => !on)}>Change dataset</button>
                <button onClick={() => setRotate(rotate => !rotate)}>Rotate</button>
                <ParentSize style={{ width: 400, height: 400 }} debounceTime={0}>
                    {parent =>
                        <SvgRoot
                            width={parent.width}
                            height={parent.height}
                            xScale={X_SCALE}
                            yScale={Y_SCALE}
                        >
                            <SvgAxisLeft/>
                            <SvgAxisBottom rotate={rotate}/>

                            <SvgContent>
                                {({ content }) =>
                                    <rect fill='hotpink' opacity={0.5} width={content.width} height={content.height}/>
                                }
                            </SvgContent>
                        </SvgRoot>
                    }
                </ParentSize>
            </div>

            <div>
                <button onClick={() => setOn(on => !on)}>Change dataset</button>
                <button onClick={() => setRotate(rotate => !rotate)}>Rotate</button>
                <ParentSize style={{ width: 400, height: 400 }} debounceTime={0}>
                    {parent =>
                        <SvgRoot
                            width={parent.width}
                            height={parent.height}
                            xScale={X_SCALE2}
                            yScale={Y_SCALE}
                        >
                            <SvgAxisLeft/>
                            <SvgAxisBottom rotate={rotate} tickFormat={tick => tick}/>

                            <SvgContent>
                                {({ content }) =>
                                    <rect fill='pink' width={content.width} height={content.height}/>
                                }
                            </SvgContent>
                        </SvgRoot>
                    }
                </ParentSize>
            </div>
        </section>
    )
}
