import * as H from 'history'
import React, { useState } from 'react'
import VisibilitySensor from 'react-visibility-sensor'
import { Observable } from 'rxjs'

import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { RepositoryFields } from '../../graphql-operations'

import { DocumentationExamplesList } from './DocumentationExamplesList'

interface Props extends SettingsCascadeProps {
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    repo: RepositoryFields
    commitID: string
    pathID: string
    count: number
}

export const DocumentationExamples: React.FunctionComponent<Props> = props => {
    const [visible, setVisible] = useState(false)
    const onVisibilityChange = (isVisible: boolean): void => {
        if (isVisible) {
            setVisible(true)
        }
    }

    return (
        <VisibilitySensor partialVisibility={true} onChange={onVisibilityChange}>
            <div className="documentation-examples mt-3 mb-3">
                {visible && <DocumentationExamplesList {...props} />}
            </div>
        </VisibilitySensor>
    )
}
