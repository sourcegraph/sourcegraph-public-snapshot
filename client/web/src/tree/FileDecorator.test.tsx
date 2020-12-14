import React from 'react'
import { render } from 'enzyme'
import { FileDecorator } from './FileDecorator'

describe('FileDecorator', () => {
    it('renders after text content', () => {
        expect(
            render(
                <FileDecorator
                    fileDecorations={[
                        { uri: 'git://github.com/test/test?branch#src', after: { contentText: 'src decoration' } },
                    ]}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot()
    })

    it('renders meter', () => {
        expect(
            render(
                <FileDecorator
                    fileDecorations={[
                        {
                            uri: 'git://github.com/test/test?branch#src',
                            meter: { value: 40, min: 0, max: 100, optimum: 70, high: 60, low: 50 },
                        },
                    ]}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot()
    })

    it('renders both after text content and meter', () => {
        expect(
            render(
                <FileDecorator
                    fileDecorations={[
                        {
                            uri: 'git://github.com/test/test?branch#src',
                            after: { contentText: 'src decoration' },
                            meter: { value: 40, min: 0, max: 100, optimum: 70, high: 60, low: 50 },
                        },
                    ]}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot()
    })

    it('respects active state', () => {
        expect(
            render(
                <FileDecorator
                    fileDecorations={[
                        { uri: 'git://github.com/test/test?branch#src', after: { contentText: 'src decoration' } },
                    ]}
                    isLightTheme={true}
                    isActive={true}
                />
            )
        ).toMatchSnapshot()
    })

    // Theming logic is already tested (fileDecorationColorForTheme())
})
