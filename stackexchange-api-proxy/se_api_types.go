package se

// Complete API docs at https://api.stackexchange.com/docs/

// These types are written to be extremely slimmed-down and use an aggressive
// filter which fetches only the bare minimum data to populate a git repository
// for the purposes of code indexing.
//
// According to the documentation filters are immutable and aggressively cached
// on StackExchange's side the filter that is used it based on the following
// API call:
//
// https://api.stackexchange.com/docs/create-filter#include=.items%3B.backoff%3B.quota_max%3B.quota_remaining%3Bquestion.body_markdown%3Banswer.body_markdown%3Bquestion.question_id%3Banswer.answer_id%3Bquestion.last_activity_date%3Banswer.last_activity_date%3Bquestion.answers%3B.error_id%3B.error_name%3B.error_message&base=none&unsafe=false&filter=default&run=true
//
// From the docs:
//
// > It is not expected that many applications will call this method
// > at runtime, filters should be pre-calculated and "baked in" in
// > the common cases. Furthermore, there are a number of built-in
// > filters which cover common use cases.

// FilterID is closely bound to the URL in the doc block
// above, do not change it without ensuring that it matches
// that which is emit by the link above.
const FilterID = "!*T0B4iEq2Y9dRWeCGhAfJ_2wl_Q2"

// Answer https://api.stackexchange.com/docs/types/answer
type Answer struct {
	LastActivityDate int    `json:"last_activity_date"`
	AnswerID         int    `json:"answer_id"`
	BodyMarkdown     string `json:"body_markdown"`
}

// Question https://api.stackexchange.com/docs/types/question
type Question struct {
	LastActivityDate int      `json:"last_activity_date"`
	QuestionID       int      `json:"question_id"`
	BodyMarkdown     string   `json:"body_markdown"`
	Answers          []Answer `json:"answers"`
}

// ResponseWrapper mirrors https://api.stackexchange.com/docs/wrapper
type QuestionsResp struct {
	Questions      []Question `json:"items"`
	QuotaMax       int        `json:"quota_max"`
	QuotaRemaining int        `json:"quota_remaining"`

	ErrorID      int    `json:"error_id"`
	ErrorName    string `json:"error_name"`
	ErrorMessage string `json:"error_message"`
}
