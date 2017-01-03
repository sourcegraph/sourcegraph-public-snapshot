import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import Event, { Emitter } from "vs/base/common/event";
import { IConfigurationServiceEvent, IConfigurationValue, getConfigurationValue } from "vs/platform/configuration/common/configuration";

import { removeWidget } from "sourcegraph/editor/authorshipWidget";
import { Features } from "sourcegraph/util/features";

const _onDidUpdateConfiguration = new Emitter<IConfigurationServiceEvent>();
const onDidUpdateConfiguration = _onDidUpdateConfiguration.event;

const config = {
	workbench: {
		quickOpen: {
			closeOnFocusLost: false,
		},
		editor: {
			enablePreview: false,
		},
	},
	explorer: {
		openEditors: {
			visible: 0,
		},
	},
	editor: {
		readOnly: true,
		automaticLayout: true,
		scrollBeyondLastLine: false,
		wrappingColumn: 0,
		fontFamily: code_font_face,
		fontSize: 15,
		lineHeight: 21,
		theme: "vs-dark",
		renderLineHighlight: "line",
		codeLens: Features.codeLens.isEnabled(),
		glyphMargin: false,
	},
};

export function enableCodeLens(): void {
	const visible = Features.codeLens.isEnabled();
	if (visible) {
		Features.codeLens.enable();
	} else {
		Features.codeLens.disable();
		removeWidget();
	}
	config.editor.codeLens = Features.codeLens.isEnabled();
	_onDidUpdateConfiguration.fire({ config } as any);
}

export class ConfigurationService {
	onDidUpdateConfiguration: Event<IConfigurationServiceEvent> = onDidUpdateConfiguration;
	getConfiguration(key: string): any {
		return config;
	}

	lookup(key: string): IConfigurationValue<any> {
		return {
			value: getConfigurationValue(config, key),
			default: getConfigurationValue(config, key),
			user: getConfigurationValue(config, key),
		};
	}

}
