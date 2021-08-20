import { storiesOf } from '@storybook/react'
import * as H from 'history'
import React, { useCallback } from 'react'
import { BehaviorSubject, of } from 'rxjs'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { StatusBarItemWithKey } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { extensionsController } from '@sourcegraph/shared/src/util/searchTestHelpers'

import webStyles from '../../SourcegraphWebApp.scss'

import { StatusBar } from './StatusBar'

const LOCATION: H.Location = { hash: '', pathname: '/', search: '', state: undefined }

const { add } = storiesOf('web/extensions/StatusBar', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container mt-3">{story()}</div>
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
            extensionsController={{
                ...extensionsController,
                extHostAPI: Promise.resolve(
                    pretendRemote<FlatExtensionHostAPI>({
                        haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
                    })
                ),
            }}
            location={LOCATION}
        />
    )
})

// TODO(tj): Carousel
