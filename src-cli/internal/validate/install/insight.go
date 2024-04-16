package install

import (
	"context"

	"github.com/sourcegraph/src-cli/internal/api"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func createInsight(ctx context.Context, client api.Client, insight Insight) (string, error) {
	var dataSeries []map[string]interface{}

	for _, ds := range insight.DataSeries {
		var series = map[string]interface{}{
			"query": ds["query"],
			"options": map[string]interface{}{
				"label":     ds["label"],
				"lineColor": ds["lineColor"],
			},
			"repositoryScope": map[string]interface{}{
				"repositories": ds["repositoryScope"],
			},
			"timeScope": map[string]interface{}{
				"stepInterval": map[string]interface{}{
					"unit":  ds["timeScopeUnit"],
					"value": ds["timeScopeValue"],
				},
			},
		}

		dataSeries = append(dataSeries, series)
	}

	q := clientQuery{
		opName: "CreateLineChartSearchInsight",
		query: `mutation CreateLineChartSearchInsight($input: LineChartSearchInsightInput!) {
			createLineChartSearchInsight(input: $input) {
	  			view {
					id
	  			}
			}
		}`,
		variables: jsonVars{
			"input": map[string]interface{}{
				"options":    map[string]interface{}{"title": insight.Title},
				"dataSeries": dataSeries,
			},
		},
	}

	var result struct {
		CreateLineChartSearchInsight struct {
			View struct {
				ID string `json:"id"`
			} `json:"view"`
		} `json:"createLineChartSearchInsight"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return "", errors.Wrap(err, "createInsight failed")
	}
	if !ok {
		return "", errors.New("createInsight failed, no data to unmarshal")
	}

	return result.CreateLineChartSearchInsight.View.ID, nil
}

func removeInsight(ctx context.Context, client api.Client, insightId string) error {
	q := clientQuery{
		opName: "DeleteInsightView",
		query: `mutation DeleteInsightView ($id: ID!) {
			deleteInsightView(id: $id){
				alwaysNil
			}
		}`,
		variables: jsonVars{
			"id": insightId,
		},
	}

	var result struct {
		Data struct {
			DeleteInsightView struct {
				AlwaysNil string `json:"alwaysNil"`
			} `json:"deleteInsightView"`
		} `json:"data"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return errors.Wrap(err, "removeInsight failed")
	}
	if !ok {
		return errors.New("removeInsight failed, no data to unmarshal")
	}

	return nil
}
