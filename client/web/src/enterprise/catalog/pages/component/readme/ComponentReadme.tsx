import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'

import { ComponentDetailFields } from '../../../../../graphql-operations'

interface Props {
    readme: NonNullable<ComponentDetailFields['readme']>
}

export const ComponentReadme: React.FunctionComponent<Props> = ({ readme }) => (
    <div className="card mb-3">
        <header className="card-header bg-transparent">
            <h4 className="card-title mb-0 font-weight-bold">
                <Link to={readme.url} className="d-flex align-items-center text-body">
                    <FileDocumentIcon className="icon-inline mr-2" /> {readme.name}
                </Link>
            </h4>
        </header>
        <Markdown dangerousInnerHTML={readme.richHTML} className="card-body p-3" />
    </div>
)
