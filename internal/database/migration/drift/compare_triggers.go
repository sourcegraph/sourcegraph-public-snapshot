package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareTriggers(actualTable, expectedTable schemas.TableDescription) []Summary {
	return compareNamedLists(actualTable.Triggers, expectedTable.Triggers, func(trigger *schemas.TriggerDescription, expectedTrigger schemas.TriggerDescription) Summary {
		createTriggerStmt := fmt.Sprintf("%s;", expectedTrigger.Definition)
		dropTriggerStmt := fmt.Sprintf("DROP TRIGGER %s ON %s;", expectedTrigger.Name, expectedTable.Name)

		if trigger == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedTrigger.Name),
				fmt.Sprintf("Missing trigger %q.%q", expectedTable.Name, expectedTrigger.Name),
				"define the trigger",
			).withStatements(createTriggerStmt)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedTrigger.Name),
			fmt.Sprintf("Unexpected properties of trigger %q.%q", expectedTable.Name, expectedTrigger.Name),
			"redefine the trigger",
		).withDiff(expectedTrigger, *trigger).withStatements(dropTriggerStmt, createTriggerStmt)
	}, func(additional []schemas.TriggerDescription) []Summary {
		summaries := []Summary{}
		for _, trigger := range additional {
			dropTriggerStmt := fmt.Sprintf("DROP TRIGGER %s ON %s;", trigger.Name, expectedTable.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, trigger.Name),
				fmt.Sprintf("Unexpected trigger %q.%q", expectedTable.Name, trigger.Name),
				"drop the trigger",
			).withStatements(dropTriggerStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	})
}
