import React, { useCallback, useEffect, useState } from 'react'

import { gql, useMutation } from '@apollo/client'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import {
    Badge,
    Button,
    Checkbox,
    Input,
    Link,
    LoadingSpinner,
    PageHeader,
    RadioButton,
    Typography,
} from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../components/MarketingBlock'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { SendJoinBetaStatsResult, SendJoinBetaStatsVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { INVALID_BETA_ID_KEY, OPEN_BETA_ID_KEY } from './NewOrganizationPage'

import styles from './JoinOpenBeta.module.scss'
interface Props extends RouteComponentProps {
    authenticatedUser: AuthenticatedUser
}

const CompanyDevsSize: string[] = [
    '1-15 developers',
    '16-25 developers',
    '26-50 developers',
    '50-100 developers',
    'More than 100 developers',
]

const CompanyRepos: string[] = [
    'GitHub.com',
    'GitHub Enterprise',
    'GitLab.com',
    'GitLab Self-Managed',
    'Bitbucket.org',
    'Other',
]

const SgUsagePlan: string[] = [
    'Understand a new codebase',
    'Fix security vulnerabilities',
    'Reuse code',
    'Respond to incidents',
    'Improve code quality',
    'Other',
]

const SEND_STATS_MUTATION = gql`
    mutation SendJoinBetaStats($stats: JSONCString!) {
        addOrgsOpenBetaStats(stats: $stats)
    }
`

export const JoinOpenBetaPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    history,
}) => {
    const [sendJoinBetaStats, { loading, error }] = useMutation<SendJoinBetaStatsResult, SendJoinBetaStatsVariables>(
        SEND_STATS_MUTATION
    )
    const [companySize, setCompanySize] = useState('')
    const [companyReposSelected, setCompanyReposSelected] = useState<string[]>([])
    const [sgUsagePlanSelected, setSgUsagePlanSelected] = useState<string[]>([])
    const [otherPlan, setOtherPlan] = useState<string | undefined>()
    const [otherRepo, setOtherRepo] = useState<string | undefined>()
    const showOtherRepo = companyReposSelected.includes('Other')
    const showOtherPlan = sgUsagePlanSelected.includes('Other')
    const otherRepoValid = !!(!showOtherRepo || otherRepo)
    const otherPlanValid = !!(!showOtherPlan || otherPlan)
    const isValidForm =
        companySize &&
        companyReposSelected.length > 0 &&
        sgUsagePlanSelected.length > 0 &&
        otherRepoValid &&
        otherPlanValid

    useEffect(() => {
        eventLogger.log(
            'CloudOpenBetaEnrollmentStarted',
            { userId: authenticatedUser.id },
            { userId: authenticatedUser.id }
        )
    }, [authenticatedUser.id])

    useEffect(() => {
        setOtherRepo(undefined)
    }, [showOtherRepo])

    useEffect(() => {
        setOtherPlan(undefined)
    }, [showOtherPlan])

    const onCompanySizeChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        setCompanySize(event.currentTarget.value)
    }

    const onCompanyRepoChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const checked = event.currentTarget.checked
        const value = event.currentTarget.value
        const selected = checked
            ? companyReposSelected.concat(value)
            : companyReposSelected.filter(item => item !== value)
        setCompanyReposSelected(selected)
    }

    const onCancelClick = (): void => {
        eventLogger.log('CloudOpenBetaEnrollmentCancelled')
        history.push(`/users/${authenticatedUser.username}/settings/organizations`)
    }

    const onSgUsagePlanChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const checked = event.currentTarget.checked
        const value = event.currentTarget.value
        const selected = checked
            ? sgUsagePlanSelected.concat(value)
            : sgUsagePlanSelected.filter(item => item !== value)
        setSgUsagePlanSelected(selected)
    }

    const onOtherPlanChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        setOtherPlan(event.currentTarget.value)
    }

    const onOtherRepoChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        setOtherRepo(event.currentTarget.value)
    }

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            eventLogger.log('OpenBetaEnrollmentContinueClicked')
            if (!event.currentTarget.checkValidity() || !isValidForm) {
                return
            }
            try {
                const stats = JSON.stringify({
                    companySize,
                    companyReposSelected,
                    sgUsagePlanSelected,
                    otherPlan: otherPlan || undefined,
                    otherRepo: otherRepo || undefined,
                })
                const result = await sendJoinBetaStats({ variables: { stats } })
                const openBetaId = result.data
                    ? (result.data as { addOrgsOpenBetaStats: string }).addOrgsOpenBetaStats
                    : INVALID_BETA_ID_KEY
                localStorage.setItem(OPEN_BETA_ID_KEY, openBetaId)
                eventLogger.log('OpenBetaEnrollmentSucceeded', { openBetaId }, { openBetaId })
                history.push(`/organizations/joinopenbeta/neworg/${openBetaId}`)
            } catch {
                eventLogger.log('OpenBetaEnrollmentFailed')
            }
        },
        [
            history,
            isValidForm,
            sendJoinBetaStats,
            companySize,
            companyReposSelected,
            sgUsagePlanSelected,
            otherPlan,
            otherRepo,
        ]
    )

    return (
        <Page className={styles.newJoinBetaPage}>
            <PageTitle title="Join Sourcegraph Open Beta" />
            <PageHeader
                path={[{ text: 'Join the open beta for Sourcegraph Cloud for small teams' }]}
                className="mb-3 mt-4"
                description={
                    <span className="text-muted">
                        Get access to Sourcegraph Cloud for small teams and start searching across your code today.
                    </span>
                }
            />
            <MarketingBlock contentClassName={styles.marketingContent}>
                <Typography.H3 className="pr-3">
                    <Badge variant="info" small={true}>
                        BETA
                    </Badge>
                </Typography.H3>
                <div>
                    Sourcegraph Cloud for small teams is in open beta. During this time, it’s free to use for 30 days.{' '}
                    <Link to="https://docs.sourcegraph.com/cloud/organizations/beta-operations">
                        Learn more about beta limitations
                    </Link>
                    .
                </div>
            </MarketingBlock>
            <Typography.H3 className="mt-4 mb-4">To get started, please tell us about your organization:</Typography.H3>
            <Form className="mb-5" onSubmit={onSubmit}>
                {error && <ErrorAlert className="mb-3" error={error} />}
                <div className={classNames('form-group', styles.formItem)}>
                    <label htmlFor="company_employees_band">About how many developers work for your company?</label>

                    <div className="mt-2">
                        {CompanyDevsSize.map(item => (
                            <div className={classNames('mb-2', styles.inputContainer)} key={item.replace(/\s/g, '_')}>
                                <RadioButton
                                    id={`cEmp_${item.replace(/\s/g, '_')}`}
                                    name="company_employees_band"
                                    value={item}
                                    checked={item === companySize}
                                    onChange={onCompanySizeChange}
                                    label={item}
                                />
                            </div>
                        ))}
                    </div>
                </div>

                <div
                    className={classNames(
                        'form-group',
                        styles.formItem,
                        showOtherRepo ? styles.otherBottom : undefined
                    )}
                >
                    <label className={styles.cbLabel} htmlFor="company_code_repo">
                        Where does your company store your code today?
                    </label>
                    <span className={classNames('text-muted d-block', styles.cbSubLabel)}>
                        <small>Select all that apply</small>
                    </span>
                    <div>
                        {CompanyRepos.map(item => (
                            <div className={classNames('mb-2', styles.inputContainer)} key={item.replace(/\s/g, '_')}>
                                <Checkbox
                                    id={`cRepo_${item.replace(/\s/g, '_')}`}
                                    name="company_code_repo"
                                    value={item}
                                    checked={companyReposSelected.includes(item)}
                                    onChange={onCompanyRepoChange}
                                    label={item}
                                />
                            </div>
                        ))}
                    </div>
                </div>

                {showOtherRepo && (
                    <div className={classNames('form-group', styles.formItem)}>
                        <Input
                            id="otherRepo_company"
                            type="text"
                            placeholder=""
                            autoCorrect="off"
                            value={otherRepo || ''}
                            label="Where else does your company store your code today?"
                            required={true}
                            onChange={onOtherRepoChange}
                            status={otherRepo === '' ? 'error' : undefined}
                        />
                    </div>
                )}

                <div
                    className={classNames(
                        'form-group',
                        styles.formItem,
                        showOtherPlan ? styles.otherBottom : undefined
                    )}
                >
                    <label className={styles.cbLabel} htmlFor="sg_usage_plan">
                        What do you plan to use Sourcegraph to do?
                    </label>
                    <span className={classNames('text-muted d-block', styles.cbSubLabel)}>
                        <small>Select all that apply</small>
                    </span>
                    <div>
                        {SgUsagePlan.map(item => (
                            <div className={classNames('mb-2', styles.inputContainer)} key={item.replace(/\s/g, '_')}>
                                <Checkbox
                                    id={`sgPlan_${item.replace(/\s/g, '_')}`}
                                    name="sg_usage_plan"
                                    value={item}
                                    checked={sgUsagePlanSelected.includes(item)}
                                    onChange={onSgUsagePlanChange}
                                    label={item}
                                />
                            </div>
                        ))}
                    </div>
                </div>
                {showOtherPlan && (
                    <div className={classNames('form-group', styles.formItem)}>
                        <Input
                            id="otherPlan_company"
                            type="text"
                            placeholder=""
                            autoCorrect="off"
                            value={otherPlan || ''}
                            label="What else do you plan to use Sourcegraph to do?"
                            required={true}
                            onChange={onOtherPlanChange}
                            status={otherPlan === '' ? 'error' : undefined}
                        />
                    </div>
                )}
                <div className={classNames('form-group d-flex justify-content-end', styles.buttonsRow)}>
                    <Button disabled={loading} variant="secondary" onClick={onCancelClick}>
                        Cancel
                    </Button>
                    <Button type="submit" disabled={loading || !isValidForm} variant="primary">
                        {loading && <LoadingSpinner />}
                        Continue
                    </Button>
                </div>
            </Form>
        </Page>
    )
}
