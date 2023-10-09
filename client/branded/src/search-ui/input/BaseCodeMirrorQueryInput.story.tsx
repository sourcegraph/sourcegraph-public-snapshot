import type { Meta, Story } from '@storybook/react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { BaseCodeMirrorQueryInput, type BaseCodeMirrorQueryInputProps } from './BaseCodeMirrorQueryInput'

const config: Meta = {
    title: 'branded/search-ui/input/BaseCodeMirrorQueryInput',
    parameters: {
        chromatic: { viewports: [500] },
    },
}

export default config

const defaultProps: BaseCodeMirrorQueryInputProps = {
    value: 'r:sourcegraph/.* test [a-z]* /is this a regex?/ author:me',
    interpretComments: false,
    patternType: SearchPatternType.standard,
}

export const Default: Story = () => (
    <BrandedStory>
        {() => (
            <>
                <div className="m-3">
                    'literal' search pattern:
                    <BaseCodeMirrorQueryInput {...defaultProps} patternType={SearchPatternType.literal} />
                </div>
                <div className="m-3">
                    'regexp' search pattern:
                    <BaseCodeMirrorQueryInput {...defaultProps} patternType={SearchPatternType.regexp} />
                </div>
                <div className="m-3">
                    'standard' search pattern:
                    <BaseCodeMirrorQueryInput {...defaultProps} patternType={SearchPatternType.standard} />
                </div>
            </>
        )}
    </BrandedStory>
)
