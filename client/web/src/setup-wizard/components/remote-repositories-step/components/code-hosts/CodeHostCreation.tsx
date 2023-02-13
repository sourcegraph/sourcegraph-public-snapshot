import { FC, ReactNode } from 'react'

import { useParams } from 'react-router-dom-v5-compat'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { Alert, H4, Link, Button } from '@sourcegraph/wildcard'

import { getCodeHostKindFromURLParam } from '../../helpers'

import { GithubConnectView } from './github/GithubConnectView'
import { CodeHostFormState } from './types'

import styles from './CodeHostCreation.module.scss'

/**
 * Renders creation UI for any supported code hosts (Github, Gitlab) based on
 * "codeHostType" URL param see root component routing logic.
 */
export const CodeHostCreation: FC = () => {
    const { codeHostType } = useParams()
    const codeHostKind = getCodeHostKindFromURLParam(codeHostType!)

    if (codeHostKind === null) {
        return (
            <Alert variant="warning">
                <H4>We either couldn't find "{codeHostType}" code host option or we do not support this</H4>
                Pick one of supported code host option <Link to="/setup/remote-repositories">here</Link>
            </Alert>
        )
    }

    // We render content inside react fragment because this view is rendered
    // within Container UI (avoid unnecessary DOM nesting)
    return (
        <CodeHostCreationView codeHostKind={codeHostKind}>
            {state => (
                <footer className={styles.footer}>
                    <Button variant="primary" size="sm" type="submit">
                        Connect
                    </Button>
                    <Button as={Link} size="sm" to="/setup/remote-repositories" variant="secondary">
                        Cancel
                    </Button>

                    {/*<small>After connecting, repositories will start sync in the background.</small>*/}
                </footer>
            )}
        </CodeHostCreationView>
    )
}

interface CodeHostCreationFormProps {
    codeHostKind: ExternalServiceKind
    children: (state: CodeHostFormState) => ReactNode
}

/**
 * Renders specific creation UI form for particular code host type. Most of
 * the code hosts have similar UI but some of them (like GitHub) have special enhanced
 * UI with pickers and other form UI.
 */
const CodeHostCreationView: FC<CodeHostCreationFormProps> = props => {
    const { codeHostKind, children } = props

    if (codeHostKind === ExternalServiceKind.GITHUB) {
        return <GithubConnectView>{children}</GithubConnectView>
    }

    // TODO: extend this view with all supported code host connections form UI
    return (
        <span>
            Oh no, we haven't implemented code host connection form for {codeHostKind}
            code host type, Todo: add this type code host connection form!
        </span>
    )
}
