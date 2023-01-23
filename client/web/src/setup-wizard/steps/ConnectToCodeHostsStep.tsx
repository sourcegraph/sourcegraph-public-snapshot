import { FC, PropsWithChildren, useState } from 'react'

import { mdiChevronDown, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { Badge, Button, Collapse, CollapseHeader, CollapsePanel, Icon, Input, Link } from '@sourcegraph/wildcard'

import { SetupStep, SetupStepActions } from '../components/SetupTabs'

import styles from './ConnectToCodeHostsStep.module.scss'

export const ConnectToCodeHostStep: FC = props => {
    const [isGitLabActive] = useState(true)

    return (
        <SetupStep>
            <section className={styles.codeHostConnections}>
                <CollapsableFormProps name="Github">
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <form>
                        <Input label="Access token" autoFocus={true} message={<GithubAccessTokenDescription />} />

                        <Button variant="primary" className="w-100 mt-3">
                            Connect GitHub
                        </Button>
                    </form>
                </CollapsableFormProps>

                <CollapsableFormProps name="GitLab" isActivated={isGitLabActive}>
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <form>
                        <Input
                            label="Access token"
                            autoFocus={true}
                            message='A GitLab access token with "api" scope. Can be a personal access token (PAT) or an OAuth token. If you are enabling permissions with identity provider type "external", this token should also have "sudo" scope.'
                        />

                        <Button variant="primary" className="w-100 mt-3">
                            Connect GitLab
                        </Button>
                    </form>
                </CollapsableFormProps>

                <CollapsableFormProps name="Bitbucket">
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <form>
                        <Input
                            label="Username"
                            autoFocus={true}
                            message='The username to use when authenticating to the Bitbucket Cloud. Also set the corresponding "appPassword" field.'
                        />

                        <Input
                            label="App password"
                            message='The app password to use when authenticating to the Bitbucket Cloud. Also set the corresponding "username" field.'
                            className="mt-2"
                        />

                        <Button variant="primary" className="w-100 mt-3">
                            Connect BitBucket
                        </Button>
                    </form>
                </CollapsableFormProps>

                <small className="mt-3">
                    Can't find your code host? We support a lot of more than just GitHub, GitLab or BitBucket, to
                    configure them visit <Link to="/site-admin/external-services">connect code hosts page.</Link>
                </small>
            </section>

            <SetupStepActions nextAvailable={true} />
        </SetupStep>
    )
}

const GithubAccessTokenDescription: FC = () => (
    <span>
        A GitHub personal access token. Create one for GitHub.com at{' '}
        <Link to="https://github.com/settings/tokens/new?description=Sourcegraph" target="_blank" rel="noopener">
            setting page
        </Link>{' '}
        (for GitHub Enterprise, replace github.com with your instance's hostname). See{' '}
        <Link
            to="https://docs.sourcegraph.com/admin/external_service/github#github-api-token-and-access"
            target="_blank"
            rel="noopener"
        >
            Sourcegraph doc page
        </Link>{' '}
        for which scopes are required for which use cases.
    </span>
)

interface CollapsableFormProps {
    name: string
    isActivated?: boolean
}

const CollapsableFormProps: FC<PropsWithChildren<CollapsableFormProps>> = props => {
    const { name, isActivated, children } = props
    const [isOpen, setOpen] = useState(false)

    return (
        <Collapse isOpen={!isActivated && isOpen} onOpenChange={setOpen}>
            <div className={styles.collapseSection}>
                <CollapseHeader
                    as={Button}
                    outline={true}
                    variant="secondary"
                    disabled={isActivated}
                    className={classNames(styles.codeHostConnectionTrigger, {
                        [styles.codeHostConnectionTriggerOpened]: isOpen,
                    })}
                >
                    <Icon aria-hidden={true} svgPath={isOpen ? mdiChevronDown : mdiChevronRight} className="mr-1" />
                    {name}
                    {isActivated && (
                        <Badge variant="success" className="ml-2">
                            Connected
                        </Badge>
                    )}
                </CollapseHeader>
                {isOpen && <CollapsePanel className="p-3">{children}</CollapsePanel>}
            </div>
        </Collapse>
    )
}
