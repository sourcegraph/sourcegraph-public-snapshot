import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { QueryBuilderPage } from './QueryBuilderPage'
import { shallow } from 'enzyme'

describe('QueryBuilderPage', () => {
    test('simple', () =>
        expect(
            shallow(<QueryBuilderPage patternType={GQL.SearchPatternType.literal} versionContext={undefined} />)
        ).toMatchSnapshot())
})
