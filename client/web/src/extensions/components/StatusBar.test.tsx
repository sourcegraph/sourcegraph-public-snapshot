import { render } from '@testing-library/react'
import * as H from 'history'
import { BehaviorSubject } from 'rxjs'

import { StatusBarItemWithKey } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { StatusBar } from './StatusBar'

const LOCATION: H.Location = { hash: '', pathname: '/', search: '', state: undefined }

describe('StatusBar', () => {
    it('renders correctly', () => {
        const getStatusBarItems = () =>
            new BehaviorSubject<StatusBarItemWithKey[]>([
                { key: 'codecov', text: 'Coverage: 96%' },
                { key: 'code-owners', text: '2 code owners', tooltip: 'Code owners: @felixbecker, @beyang' },
            ]).asObservable()

        expect(
            render(
                <StatusBar
                    getStatusBarItems={getStatusBarItems}
                    extensionsController={extensionsController}
                    location={LOCATION}
                />
            ).baseElement
        ).toMatchSnapshot()
    })

    // executes commands

    // empty state (no extensions)
})
