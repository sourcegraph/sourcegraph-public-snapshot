import * as React from 'react'

import { mdiPound } from '@mdi/js'

import { Tooltip, Icon } from '@sourcegraph/wildcard'

import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'

interface AnnotateWithCommentsProps {
    onDidUpdate: (value: boolean) => void
}

export const AnnotateWithComments: React.FunctionComponent<AnnotateWithCommentsProps> = ({ onDidUpdate }) => {
    const onClick = () => {
        onDidUpdate(true)
    }

    return (
        <Tooltip content="Annotate file with comments">
            <RepoHeaderActionButtonLink aria-label="Annotate with comments" file={false} onSelect={onClick}>
                <Icon svgPath={mdiPound} aria-hidden={true} />
            </RepoHeaderActionButtonLink>
        </Tooltip>
    )
}
