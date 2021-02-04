import React from 'react'
import { render } from '@testing-library/react'
import { StatusBar } from './StatusBar'
import { StatusBarItemWithKey } from '../../../../shared/src/api/client/api/codeEditor'
import { BehaviorSubject } from 'rxjs'

describe('StatusBar', () => {
    it('renders correctly', () => {
        const getStatusBarItems = () =>
            new BehaviorSubject<StatusBarItemWithKey[]>([
                { key: 'codecov', text: 'Coverage: 96%' },
                { key: 'code-owners', text: '2 code owners', tooltip: 'Code owners: @felixbecker, @beyang' },
            ]).asObservable()

        expect(
            render(<StatusBar getStatusBarItems={getStatusBarItems} repoName="test" commitID="test" filePath="test" />)
                .baseElement
        ).toMatchSnapshot()
    })
})
