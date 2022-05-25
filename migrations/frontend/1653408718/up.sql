ALTER TABLE survey_responses
  ADD COLUMN IF NOT EXISTS use_cases text[],
  ADD COLUMN IF NOT EXISTS other_use_case text,
  ADD COLUMN IF NOT EXISTS additional_information text;
