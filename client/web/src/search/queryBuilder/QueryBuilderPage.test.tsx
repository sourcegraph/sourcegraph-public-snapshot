import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'

import { SearchPatternType } from '../../graphql-operations'

import { QueryBuilderPage } from './QueryBuilderPage'

describe('QueryBuilderPage', () => {
    test('simple', () =>
        expect(
            createRenderer().render(
                <QueryBuilderPage
                    patternType={SearchPatternType.literal}
                    versionContext={undefined}
                    selectedSearchContextSpec="global"
                />
            )
        ).toMatchSnapshot())
})
