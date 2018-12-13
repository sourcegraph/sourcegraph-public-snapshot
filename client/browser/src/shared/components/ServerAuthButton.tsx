import * as React from 'react'
import { AuthRequiredError } from '../backend/errors'
import { DEFAULT_SOURCEGRAPH_URL, sourcegraphUrl } from '../util/context'
import { Button } from './Button'

interface Props {
    error?: AuthRequiredError
    repoPath: string
}

export class ServerAuthButton extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        const label = 'Sign in to Sourcegraph'
        const ariaLabel = 'Sign in to Sourcegraph for code intelligence on private repositories'
        const className = 'btn btn-sm tooltipped tooltipped-s aui-button'
        let url: string | undefined
        if (this.props.error) {
            url = this.props.error.url
        } else {
            url = sourcegraphUrl !== DEFAULT_SOURCEGRAPH_URL ? sourcegraphUrl : undefined
        }
        if (!url) {
            return null
        }

        return (
            <Button
                url={`${url}/${this.props.repoPath}`}
                className={className}
                ariaLabel={ariaLabel}
                label={label}
                target="_blank"
            />
        )
    }
}
