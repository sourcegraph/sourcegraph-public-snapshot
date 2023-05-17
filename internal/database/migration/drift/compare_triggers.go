package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareTriggers(actualTable, expectedTable schemas.TableDescription) []Summary {
<<<<<<< HEAD
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
=======
	return compareNamedListsStrict(
		actualTable.Triggers,
		expectedTable.Triggers,
		compareNamedListsCallbackFor(expectedTable),
		compareNamedListsAdditionalCallbackFor(expectedTable),
	)
}

func compareNamedListsCallbackFor(table schemas.TableDescription) func(_ *schemas.TriggerDescription, _ schemas.TriggerDescription) Summary {
	return func(trigger *schemas.TriggerDescription, expectedTrigger schemas.TriggerDescription) Summary {
		if trigger == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", table.GetName(), expectedTrigger.GetName()),
				fmt.Sprintf("Missing trigger %q.%q", table.GetName(), expectedTrigger.GetName()),
				"define the trigger",
			).withStatements(
				expectedTrigger.CreateStatement(),
			)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", table.GetName(), expectedTrigger.GetName()),
			fmt.Sprintf("Unexpected properties of trigger %q.%q", table.GetName(), expectedTrigger.GetName()),
			"redefine the trigger",
		).withDiff(
			expectedTrigger,
			*trigger,
		).withStatements(
			expectedTrigger.DropStatement(table),
			expectedTrigger.CreateStatement(),
		)
	}
}

func compareNamedListsAdditionalCallbackFor(table schemas.TableDescription) func(_ []schemas.TriggerDescription) []Summary {
	return func(additional []schemas.TriggerDescription) []Summary {
		summaries := []Summary{}
		for _, trigger := range additional {
			summaries = append(summaries, newDriftSummary(
				fmt.Sprintf("%q.%q", table.GetName(), trigger.GetName()),
				fmt.Sprintf("Unexpected trigger %q.%q", table.GetName(), trigger.GetName()),
				"drop the trigger",
			).withStatements(
				trigger.DropStatement(table),
			))
		}

		return summaries
	}
>>>>>>> main
}
