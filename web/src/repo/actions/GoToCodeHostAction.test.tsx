import React from 'react'
import { fireEvent, getByLabelText, getByText, render } from '@testing-library/react'
import { GoToCodeHostAction } from './GoToCodeHostAction'
import { of } from 'rxjs'
import { IGitRef } from '../../../../shared/src/graphql/schema'
import sinon from 'sinon'

describe('GoToCodeHostAction', () => {
    const mockedLocalStorage = {
        getItem: sinon.spy(() => console.log('got')),
        setItem: sinon.spy(),
    }

    beforeEach(() => {
        Object.defineProperty(window, 'localStorage', {
            value: mockedLocalStorage,
            writable: true,
        })
    })

    it('displays a popup the first time it is clicked', () => {
        const { container } = render(
            <GoToCodeHostAction
                revision="main"
                repo={{
                    externalURLs: [
                        {
                            __typename: 'ExternalLink',
                            url: '',
                            serviceType: 'github',
                        },
                    ],
                    name: 'github.com/sourcegraph/sourcegraph',
                    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                    defaultBranch: { displayName: 'main' } as IGitRef,
                }}
                browserExtensionInstalled={of(false)}
                fetchFileExternalLinks={() => {
                    console.log('called fetch file')
                    return of()
                }}
            />
        )

        // first time: popup

        sinon.assert.calledOnce(mockedLocalStorage.getItem)
        sinon.assert.notCalled(mockedLocalStorage.setItem)

        fireEvent.click(getByLabelText(container, 'View on GitHub'), { button: 0 })
        // console.log('after click', getByText(container, 'No, thanks'))

        // fireEvent.click(getByText(container, 'No, thanks'), { button: 0 })

        // sinon.assert.calledOnce(mockedLocalStorage.getItem)
        // sinon.assert.calledOnce(mockedLocalStorage.setItem)

        // fireEvent.click(getByLabelText(container, 'View on GitHub'), { button: 0 })

        // second time: no popup
        console.log('after dismissal', getByText(container, 'No, thanks'))
    })

    it('displays popup again if user clicks "Remind me later"', () => {})

    it('does not display popup if users clicks "No, thanks" or "Install"', () => {})
})
