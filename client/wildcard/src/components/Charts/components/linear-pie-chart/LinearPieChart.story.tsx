import { Meta, Story } from '@storybook/react'
import ParentSize from '@visx/responsive/lib/components/ParentSizeModern'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { LinearPieChart } from '.'

const StoryConfig: Meta = {
    title: 'wildcard/Charts',
    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        chromatic: { disableSnapshots: false, enableDarkMode: true },
    },
}
export default StoryConfig

export const LinearPieChartDemo: Story = () => {
    const data = [
        {
            language: 'JavaScript',
            value: 100,
        },
        {
            language: 'Go',
            value: 50,
        },
        {
            language: 'Python',
            value: 50,
        },
        {
            language: 'TypeScript',
            value: 100,
        },
    ]
    const getDatumColor = (datum: { language: string }): string =>
        ({
            JavaScript: 'orange',
            Go: 'teal',
            Python: 'green',
            TypeScript: 'blue',
        }[datum.language] || 'gray')
    const getDatumName = (datum: { language: string }) => datum.language
    const getDatumValue = (datum: { value: number }) => datum.value
    return (
        <ParentSize debounceTime={0}>
            {({ width, height }) => (
                <LinearPieChart
                    width={width}
                    height={height}
                    data={data}
                    getDatumColor={getDatumColor}
                    getDatumName={getDatumName}
                    getDatumValue={getDatumValue}
                    rightToLeft={true}
                >
                    {({ path, datum }) => (
                        <path fill={getDatumColor(datum)} d={path}>
                            <title>{`${getDatumName(datum)}: ${getDatumValue(datum)} LoC`}</title>
                        </path>
                    )}
                </LinearPieChart>
            )}
        </ParentSize>
    )
}
