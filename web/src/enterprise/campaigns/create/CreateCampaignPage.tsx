import React, { useCallback, useMemo, useState } from 'react'
import { PageTitle } from '../../../components/PageTitle'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsIcon } from '../icons'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import emptySample from './samples/empty.yml'
import combySample from './samples/comby.yml'
import goImportsSample from './samples/go-imports.yml'
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
    { name: 'Empty', file: emptySample },
    { name: 'Modify code using comby', file: combySample },
    { name: 'Update go imports', file: goImportsSample },
]

const sourcePreviewCommand = 'src campaign preview -f my-campaign-spec.campaign.yaml -namespace {USERNAME/ORG}'

export interface CreateCampaignPageProps extends BreadcrumbSetters {
    // Nothing for now, but using it so once this changes we get type errors in the routing files.
}

export const CreateCampaignPage: React.FunctionComponent<CreateCampaignPageProps> = ({ useBreadcrumb }) => {
    const [selectedSample, setSelectedSample] = useState<Sample>(samples[0])
    const downloadUrl = useMemo(() => 'data:text/plain;charset=utf-8,' + encodeURIComponent(selectedSample.file), [
        selectedSample,
    ])
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
            <div className="col-md-12 col-lg-10 offset-lg-1 pt-3">
                <h2>STEP 1: Write a campaign spec YAML file</h2>
                <ul className="nav nav-tabs mb-2">
                    {samples.map(sample => (
                        <SampleTabHeader
                            key={sample.name}
                            sample={sample}
                            active={selectedSample.name === sample.name}
                            setSelectedSample={setSelectedSample}
                        />
                    ))}
                    <div className="flex-grow-1 mb-1 d-flex justify-content-end">
                        <a
                            download="campaign-spec.yaml"
                            href={downloadUrl}
                            className="text-right btn btn-secondary text-nowrap"
                            data-tooltip="Download campaign-spec.yaml"
                        >
                            <FileDownloadIcon className="icon-inline" /> Download yaml
                        </a>
                    </div>
                </ul>
                <div className="create-campaign-page__specfile rounded p-3 mb-4">
                    <pre className="m-0" dangerouslySetInnerHTML={highlightedSample} />
                </div>
                <h2>STEP 2: Preview the campaign with Sourcegraph CLI</h2>
                <p className="lead">
                    Download Sourcegraph's cli tool, src-cli at{' '}
                    <a href="https://github.com/sourcegraph/src-cli" target="_blank" rel="noopener noreferrer">
                        github.com/sourcegraph/src-cli
                    </a>
                    .
                </p>
                <p className="lead">Use Sourcegraph src-cli to preview your campaign:</p>
                <div className="create-campaign-page__specfile rounded p-3 mb-3">
                    <pre className="m-0">{sourcePreviewCommand}</pre>
                </div>
            </div>
        </>
    )
}
