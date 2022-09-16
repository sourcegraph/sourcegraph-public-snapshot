package insights

func parseTimeIntervalUnit(insight searchInsight) string {
	if insight.Step.Days != nil {
		return "DAY"
	}
	if insight.Step.Hours != nil {
		return "HOUR"
	}
	if insight.Step.Weeks != nil {
		return "WEEK"
	}
	if insight.Step.Months != nil {
		return "MONTH"
	}
	if insight.Step.Years != nil {
		return "YEAR"
	}

	return ""
}

func parseTimeIntervalValue(insight searchInsight) int {
	if insight.Step.Days != nil {
		return *insight.Step.Days
	}
	if insight.Step.Hours != nil {
		return *insight.Step.Hours
	}
	if insight.Step.Weeks != nil {
		return *insight.Step.Weeks
	}
	if insight.Step.Months != nil {
		return *insight.Step.Months
	}
	if insight.Step.Years != nil {
		return *insight.Step.Years
	}

	return 1
}
