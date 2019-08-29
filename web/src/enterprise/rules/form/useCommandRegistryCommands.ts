import { useEffect, useState } from 'react'
import { CommandEntry } from '../../../../../shared/src/api/client/services/command'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'

export const useCommandRegistryCommands = (
    extensionsController: ExtensionsControllerProps['extensionsController']
): CommandEntry[] => {
    const [commands, setCommands] = useState<CommandEntry[]>(extensionsController.services.commands.commandsSnapshot)
    useEffect(() => {
        const subscription = extensionsController.services.commands.commands.subscribe(setCommands)
        return () => subscription.unsubscribe()
    }, [extensionsController.services.commands.commands])
    return commands
}
