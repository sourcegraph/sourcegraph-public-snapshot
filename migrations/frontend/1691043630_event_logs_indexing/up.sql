CREATE FUNCTION isCodyGenerationEvent(name text) RETURNS boolean
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN name = ANY(ARRAY[
    'CodyVSCodeExtension:recipe:rewrite-to-functional:executed',
    'CodyVSCodeExtension:recipe:improve-variable-names:executed',
    'CodyVSCodeExtension:recipe:replace:executed',
    'CodyVSCodeExtension:recipe:generate-docstring:executed',
    'CodyVSCodeExtension:recipe:generate-unit-test:executed',
    'CodyVSCodeExtension:recipe:rewrite-functional:executed',
    'CodyVSCodeExtension:recipe:code-refactor:executed',
    'CodyVSCodeExtension:recipe:fixup:executed',
	'CodyVSCodeExtension:recipe:translate-to-language:executed'
  ]);
END;
$$;

CREATE FUNCTION isCodyExplanationEvent(name text) RETURNS boolean
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN name = ANY(ARRAY[
    'CodyVSCodeExtension:recipe:explain-code-high-level:executed',
    'CodyVSCodeExtension:recipe:explain-code-detailed:executed',
    'CodyVSCodeExtension:recipe:find-code-smells:executed',
    'CodyVSCodeExtension:recipe:git-history:executed',
    'CodyVSCodeExtension:recipe:rate-code:executed'
  ]);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE FUNCTION isCodyActiveEvent(name text) RETURNS boolean
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN
    (name LIKE '%%cody%%' OR name LIKE '%%Cody%%')
    AND NOT
    (
    name LIKE '%%completion:started%%' OR
    name LIKE '%%completion:suggested%%' OR
    name LIKE '%%cta%%' OR
    name LIKE '%%Cta%%' OR
    name = ANY(ARRAY['CodyVSCodeExtension:CodySavedLogin:executed',
        'web:codyChat:tryOnPublicCode',
        'web:codyEditorWidget:viewed',
        'web:codyChat:pageViewed',
        'CodyConfigurationPageViewed',
        'ClickedOnTryCodySearchCTA',
        'TryCodyWebOnboardingDisplayed',
        'AboutGetCodyPopover',
        'TryCodyWeb',
        'CodySurveyToastViewed',
        'SiteAdminCodyPageViewed',
        'CodyUninstalled',
        'SpeakToACodyEngineerCTA']));
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE INDEX IF NOT EXISTS event_logs_name ON event_logs USING GIN (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS event_logs_name_is_cody_explanation_event ON event_logs USING btree (isCodyExplanationEvent(name));

CREATE INDEX IF NOT EXISTS event_logs_name_is_cody_generation_event ON event_logs USING btree (isCodyGenerationEvent(name));

CREATE INDEX IF NOT EXISTS event_logs_name_is_cody_active_event ON event_logs USING btree (isCodyActiveEvent(name));
