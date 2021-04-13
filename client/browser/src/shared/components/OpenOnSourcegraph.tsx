import classNames from 'classnames'
import * as React from 'react'

import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { OpenInSourcegraphProps } from '../repo'
import { getPlatformName } from '../util/context'

import { SourcegraphIconButton, SourcegraphIconButtonProps } from './SourcegraphIconButton'

interface Props extends SourcegraphIconButtonProps {
    openProps: OpenInSourcegraphProps
}

export const OpenOnSourcegraph: React.FunctionComponent<Props> = ({
    openProps: { sourcegraphURL, repoName, revision, filePath },
    className,
    ...props
}) => {
    const url = new URL(toPrettyBlobURL({ repoName, revision, filePath }), sourcegraphURL)
    url.searchParams.set('utm_source', getPlatformName())
    return <SourcegraphIconButton {...props} className={classNames('open-on-sourcegraph', className)} href={url.href} />
}
