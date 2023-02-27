import * as vscode from 'vscode'

const CODY_ENDPOINT = 'cody.sgdev.org'

export type ConfigurationUseContext = 'embeddings' | 'keyword' | 'none'

export interface Configuration {
	enable: boolean
	serverEndpoint: string
	embeddingsEndpoint: string
	codebase?: string
	debug: boolean
	useContext: ConfigurationUseContext
	experimentalSuggest: boolean
}

export function getConfiguration(config: vscode.WorkspaceConfiguration): Configuration {
	return {
		enable: config.get('sourcegraph.cody.enable', true),
		serverEndpoint: config.get('cody.serverEndpoint')!,
		embeddingsEndpoint: config.get('cody.embeddingsEndpoint')!,
		codebase: config.get('cody.codebase'),
		debug: config.get('cody.debug', false),
		useContext: config.get<ConfigurationUseContext>('cody.useContext') || 'embeddings',
		experimentalSuggest: config.get('cody.experimental.suggest', false),
	}
}
