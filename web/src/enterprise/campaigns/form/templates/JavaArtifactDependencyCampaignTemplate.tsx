import LanguageJavaIcon from 'mdi-react/LanguageJavaIcon'
import React from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'

interface Props extends CampaignTemplateComponentContext {}

const JavaArtifactDependencyCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const JavaArtifactDependencyCampaignTemplate: CampaignTemplate = {
    id: 'javaArtifactDependency',
    title: 'Java artifact dependency deprecation/ban',
    detail:
        'Deprecate or ban a Java dependency in artifacts and Maven/Gradle build configs, opening issues/changesets for all affected code owners.',
    icon: LanguageJavaIcon,
    renderForm: JavaArtifactDependencyCampaignTemplateForm,
}
