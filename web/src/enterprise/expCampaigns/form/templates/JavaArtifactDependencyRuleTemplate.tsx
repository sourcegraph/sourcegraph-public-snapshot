import LanguageJavaIcon from 'mdi-react/LanguageJavaIcon'
import React from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '../templates'

interface Props extends RuleTemplateComponentContext {}

const JavaArtifactDependencyCampaignTemplateForm: React.FunctionComponent<Props> = ({}) => <p>hello world</p>

export const JavaArtifactDependencyRuleTemplate: RuleTemplate = {
    id: 'javaArtifactDependency',
    title: 'Java artifact dependency deprecation/ban',
    detail:
        'Deprecate or ban a Java dependency in artifacts and Maven/Gradle build configs, opening issues/changesets for all affected code owners.',
    icon: LanguageJavaIcon,
    renderForm: JavaArtifactDependencyCampaignTemplateForm,
}
