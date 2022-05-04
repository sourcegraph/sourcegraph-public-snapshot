ALTER TABLE survey_responses
  DROP COLUMN IF EXISTS use_cases,
  DROP COLUMN IF EXISTS additional_information;
