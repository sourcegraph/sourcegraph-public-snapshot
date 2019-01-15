package se

// Complete API docs at https://api.stackexchange.com/docs/

// User mirrors https://api.stackexchange.com/docs/types/user
type User struct {
	Reputation   int    `json:"reputation"`
	UserID       int    `json:"user_id"`
	UserType     string `json:"user_type"`
	AcceptRate   int    `json:"accept_rate"`
	ProfileImage string `json:"profile_image"`
	DisplayName  string `json:"display_name"`
	Link         string `json:"link"`
}

// Answer mirrors https://api.stackexchange.com/docs/types/answer
type Answer struct {
	Owner            User   `json:"owner"`
	IsAccepted       bool   `json:"is_accepted"`
	Score            int    `json:"score"`
	LastActivityDate int    `json:"last_activity_date"`
	CreationDate     int    `json:"creation_date"`
	AnswerID         int    `json:"answer_id"`
	QuestionID       int    `json:"question_id"`
	Body             string `json:"body"`
}

// Question mirrors https://api.stackexchange.com/docs/types/question
type Question struct {
	Tags             []string `json:"tags"`
	Owner            User     `json:"owner"`
	IsAnswered       bool     `json:"is_answered"`
	ViewCount        int      `json:"view_count"`
	AcceptedAnswerID int      `json:"accepted_answer_id"`
	AnswerCount      int      `json:"answer_count"`
	Score            int      `json:"score"`
	LastActivityDate int      `json:"last_activity_date"`
	CreationDate     int      `json:"creation_date"`
	QuestionID       int      `json:"question_id"`
	Link             string   `json:"link"`
	Title            string   `json:"title"`
}

// AnswersResp mirrors https://api.stackexchange.com/docs/answers
type AnswersResp struct {
	Items          []Answer `json:"items"`
	HasMore        bool     `json:"has_more"`
	QuotaMax       int      `json:"quota_max"`
	QuotaRemaining int      `json:"quota_remaining"`
}

// QuestionsResp mirrors https://api.stackexchange.com/docs/questions
type QuestionsResp struct {
	Questions      []Question `json:"items"`
	HasMore        bool       `json:"has_more"`
	QuotaMax       int        `json:"quota_max"`
	QuotaRemaining int        `json:"quota_remaining"`
}
