import type React from 'react'
import { useEffect } from 'react'

import classNames from 'classnames'
import { useNavigate, useParams } from 'react-router-dom'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ButtonLink, Card, AnchorLink, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { CodyProRoutes } from '../codyProRoutes'

import styles from './CodySwitchAccountPage.module.scss'

interface CodySwitchAccountPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
}

export const CodySwitchAccountPage: React.FunctionComponent<CodySwitchAccountPageProps> = ({
    authenticatedUser,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.switch-account', 'view')
    }, [telemetryRecorder])

    const navigate = useNavigate()
    const { username = '' } = useParams()

    const accountSwitchNotRequired = !username || !authenticatedUser || authenticatedUser.username === username
    useEffect(() => {
        if (accountSwitchNotRequired) {
            navigate(CodyProRoutes.Manage)
        }
    }, [accountSwitchNotRequired, navigate])

    if (accountSwitchNotRequired || !authenticatedUser) {
        return null
    }

    return (
        <Page className="d-flex flex-column">
            <PageTitle title="Switch Account" />
            <div className="flex-1" />
            <Card className={classNames('d-flex flex-column flex-1 mx-auto p-4', styles.switchAccountCard)}>
                <Text>
                    Your Cody client is signed in with a different account <strong>(@{username})</strong>. To manage the
                    account being used by your Cody client, sign out and sign in with the intended account.
                </Text>
                <Button to="/-/sign-out" as={AnchorLink} variant="primary" className="mt-3">
                    Sign out to switch accounts
                </Button>
                <div className="my-4 d-flex align-items-center justify-content-center">
                    <hr className="flex-1" />
                    <Text className="text-muted mb-0 px-2" size="small">
                        or
                    </Text>
                    <hr className="flex-1" />
                </div>

                <div className="d-flex align-items-center border rounded p-2 mb-2">
                    <div>
                        <UserAvatar size={32} user={authenticatedUser} />
                    </div>
                    <div className="d-flex flex-column ml-2">
                        <Text className="mb-0" weight="medium">
                            {authenticatedUser.displayName} (@{authenticatedUser.username})
                        </Text>
                        <Text className="mb-0 text-muted">{authenticatedUser.emails[0].email}</Text>
                    </div>
                </div>
                <ButtonLink to={CodyProRoutes.Manage} variant="secondary">
                    Continue
                </ButtonLink>
            </Card>
            <div className="flex-1" />
        </Page>
    )
}
