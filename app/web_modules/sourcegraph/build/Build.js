import React from "react";

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
	return "default";
}

export function taskClass(task) {
	if (task.Success && !task.Skipped) return "success";
	if (task.Failure && !task.Skippe) return "danger";
	if (task.Warnings) return "warning";
	if (!task.Success && !task.Failure && !task.Skipped) return "info";
	return "default";
}

export function elapsed(buildOrTask) {
	if (!buildOrTask.StartedAt) return null;
	return (
		<div>
			{formatDurationHHMMSS(new Date(buildOrTask.EndedAt).getTime() - new Date(buildOrTask.StartedAt).getTime())}
		</div>
	);
}

// formatDurationHHMMSS returns a string like "9:03" (meaning 9
// minutes, 3 seconds) given a number of milliseconds.
export function formatDurationHHMMSS(ms) {
	let s = Math.floor(ms / 1000);
	let m = Math.floor(s / 60);
	let h = Math.floor(m / 60);

	let str = "";

	if (h > 0) {
		str += `${h}:`;
	}

	if (h > 0) {
		str += `0${m}:`;
	} else {
		str += `${m}:`;
	}

	if (s >= 10) {
		str += s;
	} else {
		str += `0${s}`;
	}

	return str;
}

