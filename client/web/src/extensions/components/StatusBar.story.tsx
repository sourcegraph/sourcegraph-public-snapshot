import { useCallback } from 'react'

import { storiesOf } from '@storybook/react'
import * as H from 'history'
import { BehaviorSubject, of } from 'rxjs'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { StatusBarItemWithKey } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { AppRouterContainer } from '../../components/AppRouterContainer'

import { StatusBar } from './StatusBar'

import webStyles from '../../SourcegraphWebApp.scss'

const LOCATION: H.Location = { hash: '', pathname: '/', search: '', state: undefined }

const { add } = storiesOf('web/extensions/StatusBar', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <AppRouterContainer>
            <div className="container mt-3">{story()}</div>
        </AppRouterContainer>
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
