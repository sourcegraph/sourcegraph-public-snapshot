import * as React from 'react'

import classNames from 'classnames'

import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import type { OpenInSourcegraphProps } from '../repo'

import { SourcegraphIconButton, type SourcegraphIconButtonProps } from './SourcegraphIconButton'

interface Props extends SourcegraphIconButtonProps {
    openProps: OpenInSourcegraphProps
}

export const OpenOnSourcegraph: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    openProps: { sourcegraphURL, repoName, revision, filePath },
    className,
    ...props
}) => {
    const url = new URL(toPrettyBlobURL({ repoName, revision, filePath }), sourcegraphURL)
    return (
        <SourcegraphIconButton
            {...props}
            className={classNames('open-on-sourcegraph', className)}
            dataTestId="open-on-sourcegraph"
            href={url.href}
        />
    )
}
