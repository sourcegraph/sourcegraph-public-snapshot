import { Registry } from "vs/platform/platform";
import { Extensions, IStatusbarRegistry, StatusbarAlignment, StatusbarItemDescriptor } from "vs/workbench/browser/parts/statusbar/statusbar";

import { FeedbackStatusbarItem } from "sourcegraph/feedback/feedbackStatusbarItem";
import { Features } from "sourcegraph/util/features";
import { updateConfiguration } from "sourcegraph/workbench/ConfigurationService";

if (Features.beta.isEnabled()) {
	updateConfiguration((config: any) => {
		// Show status bar because the feedback button is on the status bar.
		config.workbench.statusBar.visible = true;
	});

	// Register feedback item in the status bar.
	Registry.as<IStatusbarRegistry>(Extensions.Statusbar).registerStatusbarItem(new StatusbarItemDescriptor(
		FeedbackStatusbarItem,
		StatusbarAlignment.RIGHT,
		-100 /* Low Priority */
	));
}
