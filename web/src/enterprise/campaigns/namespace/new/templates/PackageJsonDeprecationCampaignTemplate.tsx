import PackageIcon from 'mdi-react/PackageIcon'
import React from 'react'
import { CampaignTemplate, CampaignTemplateContext } from '.'

interface Props extends CampaignTemplateContext {}

const PackageJsonDeprecationCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const PackageJsonDeprecationCampaignTemplate: CampaignTemplate = {
    id: 'deprecatePackageJsonDependency',
    title: 'Deprecate a package.json dependency',
    detail:
        'Find packages with the deprecated dependency and/or version range and open issues/changesets to remove the dependency.',
    icon: PackageIcon,
    renderForm: PackageJsonDeprecationCampaignTemplateForm,
}
