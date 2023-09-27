pbckbge insights

func pbrseTimeIntervblUnit(insight sebrchInsight) string {
	if insight.Step.Dbys != nil {
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
	if insight.Step.Yebrs != nil {
		return "YEAR"
	}

	return ""
}

func pbrseTimeIntervblVblue(insight sebrchInsight) int {
	if insight.Step.Dbys != nil {
		return *insight.Step.Dbys
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
	if insight.Step.Yebrs != nil {
		return *insight.Step.Yebrs
	}

	return 1
}
