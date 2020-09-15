import React, { useCallback, useMemo, useState } from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsIcon } from '../icons'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { AuthenticatedUser } from '../../../auth'
import helloWorldSample from './samples/empty.campaign.yaml'
import combySample from './samples/comby.campaign.yaml'
import goImportsSample from './samples/go-imports.campaign.yaml'
import minimalSample from './samples/minimal.campaign.yaml'
import classNames from 'classnames'
import { highlightCodeSafe } from '../../../../../shared/src/util/markdown'

interface SampleTabHeaderProps {
    sample: Sample
    active: boolean
    setSelectedSample: (sample: Sample) => void
}

const SampleTabHeader: React.FunctionComponent<SampleTabHeaderProps> = ({ sample, active, setSelectedSample }) => {
    const onClick = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedSample(sample)
        },
        [setSelectedSample, sample]
    )
    return (
        <li className="nav-item">
            <a href="" onClick={onClick} className={classNames('nav-link', active && 'active')}>
                {sample.name}
            </a>
        </li>
    )
}

interface Sample {
    name: string
    file: string
}

const samples: Sample[] = [
    { name: 'hello-world.campaign.yaml', file: helloWorldSample },
    { name: 'modify-with-comby.campaign.yaml', file: combySample },
    { name: 'update-go-imports.campaign.yaml', file: goImportsSample },
    { name: 'minimal.campaign.yaml', file: minimalSample },
]

export interface CreateCampaignPageProps extends BreadcrumbSetters {
    authenticatedUser: Pick<AuthenticatedUser, 'username'> | null
}

export const CreateCampaignPage: React.FunctionComponent<CreateCampaignPageProps> = ({
    authenticatedUser,
    useBreadcrumb,
}) => {
    const [selectedSample, setSelectedSample] = useState<Sample>(samples[0])
    const highlightedSample = useMemo(() => ({ __html: highlightCodeSafe(selectedSample.file, 'yaml') }), [
        selectedSample.file,
    ])
    useBreadcrumb(useMemo(() => ({ element: <>Create campaign</>, key: 'createCampaignPage' }), []))
    return (
        <>
            <PageTitle title="Create campaign" />
            <PageHeader
                icon={CampaignsIcon}
                title={
                    <>
                        Create campaign{' '}
                        <sup>
                            <span className="badge badge-merged text-uppercase">Beta</span>
                        </sup>
                    </>
                }
            />
            <div className="pt-3">
                <h2>1. Write a campaign spec YAML file</h2>
                <p>
                    The campaign spec (
                    <a
                        href="https://docs.sourcegraph.com/user/campaigns#campaign-specs"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        syntax reference
                    </a>
                    ) describes what the campaign does. You'll provide it when previewing, creating, and updating
                    campaigns. We recommend committing it to source control.
                </p>
                <h4>Examples:</h4>
                <ul className="nav nav-pills mb-2">
                    {samples.map(sample => (
                        <SampleTabHeader
                            key={sample.name}
                            sample={sample}
                            active={selectedSample.name === sample.name}
                            setSelectedSample={setSelectedSample}
                        />
                    ))}
                </ul>
                <div className="p-3 mb-4 border">
                    <pre className="m-0" dangerouslySetInnerHTML={highlightedSample} />
                </div>
                <h2>2. Preview the campaign with Sourcegraph CLI</h2>
                <p>
                    Use the{' '}
                    <a href="https://github.com/sourcegraph/src-cli" rel="noopener noreferrer" target="_blank">
                        Sourcegraph CLI (src)
                    </a>{' '}
                    to preview the commits and changesets that your campaign will make:
                </p>
                <pre className="">
                    <code>
                        src campaign preview -namespace {authenticatedUser ? authenticatedUser.username : 'NAMESPACE'}{' '}
                        -f {selectedSample.name}
                    </code>
                </pre>
                <p>
                    Follow the URL printed in your terminal to see the preview and (when you're ready) create the
                    campaign.
                </p>
                <hr className="mt-5" />
                <p className="mt-2 text-muted">
                    Want more help? See{' '}
                    <a href="/help/user/campaigns" rel="noopener noreferrer" target="_blank">
                        campaigns documentation
                    </a>
                    .
                </p>
            </div>
        </>
    )
}
