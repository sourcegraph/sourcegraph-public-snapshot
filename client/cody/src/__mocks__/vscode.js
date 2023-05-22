module.exports = {
  window: {
    showInformationMessage: () => undefined,
    showErrorMessage: () => undefined,
    showWarningMessage: () => undefined,
    showQuickPick: () => undefined,
    showInputBox: () => undefined,
    createOutputChannel: () => undefined,
  },
  getConfiguration: () => undefined,
  WorkspaceConfiguration: {
    get: () => undefined,
    update: () => undefined,
  },
  workspace: {
    getConfiguration: () => undefined,
  },
  ConfigurationTarget: {
    Global: undefined,
  },
}
