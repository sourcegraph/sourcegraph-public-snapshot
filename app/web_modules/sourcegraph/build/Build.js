import classNames from "classnames";
import moment from "moment";
import React from "react";

// buildStatus returns a textual status description for the build.
export function buildStatus(b) {
	if (b.Killed) {
		return "Killed";
	}
	if (b.Warnings) {
		return "Warnings";
	}
	if (b.Failure) {
		return "Failed";
	}
	if (b.Success) {
		return "Succeeded";
	}
	if (b.StartedAt && !b.EndedAt) {
		return "Started";
	}
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
	if (task.Skipped) {
		return {icon: "fa fa-ban", text: "text-muted", listGroupItem: ""};
	}
	if (!task.StartedAt) {
		return {icon: "fa fa-circle-o-notch", text: "", listGroupItem: ""};
	}
	if (task.Warnings) {
		return {icon: "fa fa-exclamation-circle", text: "text-warning", listGroupItem: "list-group-item-warning"};
	}
	if (task.Success) {
		return {icon: "fa fa-check", text: "text-success", listGroupItem: "list-group-item-success"};
	}
	if (task.Failure) {
		return {icon: "fa fa-exclamation-circle", text: "text-danger", listGroupItem: "list-group-item-danger"};
	}
	return {icon: "fa fa-circle-o-notch fa-spin", text: "text-info", listGroupItem: "list-group-item-info"};
}

export function elapsed(buildOrTask) {
	if (!buildOrTask.StartedAt) return null;
	return (
		<div className="elapsed">
			<i className="fa fa-clock-o"></i>&nbsp;
			{formatDurationHHMMSS(moment.duration(moment(buildOrTask.EndedAt).diff(buildOrTask.StartedAt)))}
		</div>
	);
}

// formatDurationHHMMSS returns a string like "9:03" (meaning 9
// minutes, 3 seconds) given a moment duration value.
export function formatDurationHHMMSS(d) {
	let h = Math.floor(d.asHours());
	let m = d.minutes();
	let s = d.seconds();

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

export function panelClass(buildOrTask) {
	return classNames({
		"panel": true,
		"panel-default": true,
		"panel-success": buildOrTask.Success && !buildOrTask.Skipped,
		"panel-danger": buildOrTask.Failure && !buildOrTask.Skipped,
		"panel-warning": buildOrTask.Warnings,
		"panel-info": !buildOrTask.Success && !buildOrTask.Failure && !buildOrTask.Skipped,
	});
}
