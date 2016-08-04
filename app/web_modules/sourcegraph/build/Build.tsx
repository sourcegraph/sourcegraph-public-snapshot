// tslint:disable

import * as React from "react";
import {formatDuration} from "sourcegraph/util/TimeAgo";

export function updatedAt(b) {
	return b.EndedAt || b.StartedAt || b.CreatedAt || null;
}

// buildStatus returns a textual status description for the build.
export function buildStatus(b) {
	if (b.Killed) return "Killed";
	if (b.Warnings) return "Warnings";
	if (b.Failure) return "Failed";
	if (b.Success) return "Succeeded";
	if (b.StartedAt && !b.EndedAt) return "Started";
	return "Queued";
}

// buildClass returns the CSS class for the build.
export function buildClass(b) {
	switch (buildStatus(b)) {
	case "Killed":
		return "danger";
	case "Warnings":
		return "warning";
	case "Failed":
		return "danger";
	case "Succeeded":
		return "success";
	case "Started":
		return "info";
	}
	return "normal";
}

export function taskClass(task) {
	if (task.Warnings) return "warning";
	if (task.Failure && !task.Skipped) return "danger";
	if (!task.Success && !task.Failure && !task.Skipped) return "info";
	if (task.Success && !task.Skipped) return "success";
	return "normal";
}

export function elapsed(buildOrTask) {
	if (!buildOrTask.StartedAt) return null;
	return (
		<div>
			{formatDuration((buildOrTask.EndedAt ? new Date(buildOrTask.EndedAt).getTime() : Date.now()) - new Date(buildOrTask.StartedAt).getTime())}
		</div>
	);
}

// E.g., extract "master" from "master~10".
export function guessBranchName(rev) {
	if (!rev) return null;
	if (rev.length === 40) return null;
	return rev.replace(/[~^].*$/, "");
}
