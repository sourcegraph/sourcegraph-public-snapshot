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

export type { StepConfiguration, StepComponentProps } from './setup-steps'

export type { LocalCodeHost } from './local-repositories-step'
export { callFilePicker, useLocalExternalServices } from './local-repositories-step'

export { RemoteRepositoriesStep } from './remote-repositories-step'
export { SyncRepositoriesStep } from './SyncRepositoriesStep'
