import React from 'react'
import { ViewPage } from './ViewPage'
import * as H from 'history'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import { Controller } from '../../../shared/src/extensions/controller'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { of } from 'rxjs'
import { mount } from 'enzyme'

jest.mock('@sourcegraph/react-loading-spinner', () => ({ LoadingSpinner: 'LoadingSpinner' }))
jest.mock('./QueryInputInViewContent', () => ({ QueryInputInViewContent: 'QueryInputInViewContent' }))

const commonProps: Omit<React.ComponentProps<typeof ViewPage>, 'viewID' | 'extraPath' | '_getView'> = {
    settingsCascade: { final: {}, subjects: [] },
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    setCaseSensitivity: () => undefined,
    setPatternType: () => undefined,
    history: H.createMemoryHistory(),
    location: H.createLocation('/'),
    extensionsController: { services: { contribution: { getContributions: () => ({}) } } } as Controller,
    copyQueryButton: false,
    versionContext: undefined,
}

describe('ViewPage', () => {
    test('view is loading', () => {
        expect(
            mount(<ViewPage {...commonProps} viewID="v" extraPath="" _getView={() => of(undefined)} />).children()
        ).toMatchSnapshot()
    })

    test('view not found', () => {
        expect(
            mount(<ViewPage {...commonProps} viewID="v" extraPath="" _getView={() => of(null)} />).children()
        ).toMatchSnapshot()
    })

    test('renders view', () => {
        expect(
            mount(
                <ViewPage
                    {...commonProps}
                    viewID="v"
                    extraPath=""
                    _getView={() =>
                        of({
                            title: 't',
                            content: [
                                {
                                    kind: MarkupKind.Markdown,
                                    value: '**a**',
                                },
                                {
                                    kind: MarkupKind.PlainText,
                                    value: '*b*',
                                },
                                {
                                    value: '*c*',
                                },
                                {
                                    component: 'QueryInput',
                                    props: { implicitQueryPrefix: 'x' },
                                },
                            ],
                        })
                    }
                />
            ).children()
        ).toMatchSnapshot()
    })
})
