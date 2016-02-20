package gochimp

import (
	"errors"
)

const subaccounts_list_endpoint string = "/subaccounts/list.json"
const subaccounts_add_endpoint string = "/subaccounts/add.json"
const subaccounts_info_endpoint string = "/subaccounts/info.json"
const subaccounts_update_endpoint string = "/subaccounts/update.json"
const subaccounts_delete_endpoint string = "/subaccounts/delete.json"
const subaccounts_pause_endpoint string = "/subaccounts/pause.json"
const subaccounts_resume_endpoint string = "/subaccounts/resume.json"

type SubaccountTimeSeries struct {
	Sent         int32 `json:"sent"`
	HardBounces  int32 `json:"hard_bounces"`
	SoftBounces  int32 `json:"soft_bounces"`
	Rejects      int32 `json:"rejects"`
	Complaints   int32 `json:"complaints"`
	Unsubs       int32 `json:"unsubs"`
	Opens        int32 `json:"opens"`
	UniqueOpens  int32 `json:"unique_opens"`
	Clicks       int32 `json:"clicks"`
	UniqueClicks int32 `json:"unique_clicks"`
}

type SubaccountInfo struct {
	Id          string               `json:"id"`
	Name        string               `json:"name"`
	Notes       string               `json:"notes"`
	CustomQuota int32                `json:"custom_quota"`
	Status      string               `json:"status"`
	Reputation  int32                `json:"reputation"`
	CreatedAt   APITime              `json:"created_at"`
	FirstSentAt APITime              `json:"first_sent_at"`
	SentWeekly  int32                `json:"sent_weekly"`
	SentMonthly int32                `json:"sent_monthly"`
	SentTotal   int32                `json:"sent_total"`
	SentHourly  int32                `json:"sent_hourly"`
	HourlyQuota int32                `json:"hourly_quota"`
	Last30Days  SubaccountTimeSeries `json:"last_30_days"`
}

func (a *MandrillAPI) SubaccountList() (response []SubaccountInfo, err error) {
	params := make(map[string]interface{})
	err = parseMandrillJson(a, subaccounts_list_endpoint, params, &response)
	return
}

func (a *MandrillAPI) SubaccountAdd(id string, name string, notes string, custom_quota int32) (response SubaccountInfo, err error) {
	return a.subaccountCreateOrUpdate(id, subaccounts_add_endpoint, name, notes, custom_quota)
}

func (a *MandrillAPI) SubaccountInfo(id string) (response SubaccountInfo, err error) {
	return a.subaccountSimpleInteraction(id, subaccounts_info_endpoint)
}

func (a *MandrillAPI) SubaccountUpdate(id string, name string, notes string, custom_quota int32) (response SubaccountInfo, err error) {
	return a.subaccountCreateOrUpdate(id, subaccounts_update_endpoint, name, notes, custom_quota)
}

func (a *MandrillAPI) SubaccountDelete(id string) (response SubaccountInfo, err error) {
	return a.subaccountSimpleInteraction(id, subaccounts_delete_endpoint)
}

func (a *MandrillAPI) SubaccountPause(id string) (response SubaccountInfo, err error) {
	return a.subaccountSimpleInteraction(id, subaccounts_pause_endpoint)
}

func (a *MandrillAPI) SubaccountResume(id string) (response SubaccountInfo, err error) {
	return a.subaccountSimpleInteraction(id, subaccounts_resume_endpoint)
}

func (a *MandrillAPI) subaccountSimpleInteraction(id, endpoint string) (response SubaccountInfo, err error) {
	if id == "" {
		return response, errors.New("id cannot be blank")
	}

	params := map[string]interface{}{"id": id}
	err = parseMandrillJson(a, endpoint, params, &response)
	return
}

func (a *MandrillAPI) subaccountCreateOrUpdate(id, endpoint string, name string, notes string, custom_quota int32) (response SubaccountInfo, err error) {
	if id == "" {
		return response, errors.New("id cannot be blank")
	}

	if len(id) > 255 {
		return response, errors.New("id cannot be over 255 characters")
	}

	if len(name) > 1024 {
		return response, errors.New("name cannot be over 1024 characters")
	}

	params := map[string]interface{}{
		"id":    id,
		"name":  name,
		"notes": notes,
	}

	if custom_quota > 0 {
		params["custom_quota"] = custom_quota
	}

	err = parseMandrillJson(a, endpoint, params, &response)
	return

}
