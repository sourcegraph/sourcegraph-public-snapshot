import * as React from 'react'

import classNames from 'classnames'

import { createURLWithUTM } from '@sourcegraph/shared/src/tracking/utm'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { OpenInSourcegraphProps } from '../repo'
import { getPlatformName } from '../util/context'

import { SourcegraphIconButton, SourcegraphIconButtonProps } from './SourcegraphIconButton'

interface Props extends SourcegraphIconButtonProps {
    openProps: OpenInSourcegraphProps
}

export const OpenOnSourcegraph: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    openProps: { sourcegraphURL, repoName, revision, filePath },
    className,
    ...props
}) => {
    const url = createURLWithUTM(new URL(toPrettyBlobURL({ repoName, revision, filePath }), sourcegraphURL), {
        utm_source: getPlatformName(),
        utm_campaign: 'open-on-sourcegraph',
    })
    return (
        <SourcegraphIconButton
            {...props}
            className={classNames('open-on-sourcegraph', className)}
            dataTestId="open-on-sourcegraph"
            href={url.href}
        />
    )
}
