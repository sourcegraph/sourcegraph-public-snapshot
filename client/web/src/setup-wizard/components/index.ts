// Re-export setup steps to reuse them in App/Cody settings pages.

export {
    SetupStepsHeader,
    SetupStepsContent,
    SetupStepsFooter,
    SetupStepsRoot,
    FooterWidget,
    CustomNextButton,
    SetupStepsContext,
} from './setup-steps'

export type { StepConfiguration } from './setup-steps'

export {
    LocalRepositoriesStep,
    callFilePicker,
    useLocalRepositories,
    useLocalRepositoriesPaths,
    useNewLocalRepositoriesPaths,
} from './local-repositories-step'

export { RemoteRepositoriesStep } from './remote-repositories-step'
export { SyncRepositoriesStep } from './SyncRepositoriesStep'
