import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import * as GQL from '../../../../shared/src/graphql/schema'
import { QueryBuilderPage } from './QueryBuilderPage'

describe('QueryBuilderPage', () => {
    test('simple', () =>
        expect(
            createRenderer().render(
                <QueryBuilderPage patternType={GQL.SearchPatternType.literal} versionContext={undefined} />
            )
        ).toMatchSnapshot())
})
