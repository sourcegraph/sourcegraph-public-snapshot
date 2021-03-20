import React, { useCallback } from 'react'
import { storiesOf } from '@storybook/react'
import webStyles from '../../SourcegraphWebApp.scss'
import { StatusBar } from './StatusBar'
import { BehaviorSubject } from 'rxjs'
import { extensionsController } from '../../../../shared/src/util/searchTestHelpers'
import * as H from 'history'
import { StatusBarItemWithKey } from '../../../../shared/src/api/extension/api/codeEditor'

const LOCATION: H.Location = { hash: '', pathname: '/', search: '', state: undefined }

const { add } = storiesOf('web/extensions/StatusBar', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('two items', () => {
    const getStatusBarItems = useCallback(
        () =>
            new BehaviorSubject<StatusBarItemWithKey[]>([
                { key: 'codecov', text: 'Coverage: 96%' },
                { key: 'code-owners', text: '2 code owners', tooltip: 'Code owners: @felixbecker, @beyang' },
            ]).asObservable(),
        []
    )
    return (
        <StatusBar
            getStatusBarItems={getStatusBarItems}
            extensionsController={extensionsController}
            location={LOCATION}
        />
    )
})

// TODO(tj): Carousel
