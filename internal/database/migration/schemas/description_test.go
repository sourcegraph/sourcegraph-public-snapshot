package schemas

import "testing"

func TestNormalizeFunction(t *testing.T) {
	for _, testCase := range []struct {
		name string
		lhs  string
		rhs  string
	}{
		{
			name: "equivalent",
			lhs: `
				CREATE OR REPLACE FUNCTION public.lsif_data_docs_search_private_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
			rhs: `
				CREATE OR REPLACE FUNCTION public.lsif_data_docs_search_private_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
		},
		{
			name: "different spacing",
			lhs: `
				CREATE OR REPLACE FUNCTION
				public.lsif_data_docs_search_private_delete()
				RETURNS trigger
				LANGUAGE plpgsql
				AS $function$
				BEGIN
					UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
			rhs: `
				CREATE OR REPLACE FUNCTION public.lsif_data_docs_search_private_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
		},
		{
			name: "comments differ",
			lhs: `
				CREATE OR REPLACE FUNCTION public.lsif_data_docs_search_private_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
					-- should not matter that this is here!
				END $function$;
			`,
			rhs: `
				CREATE OR REPLACE FUNCTION public.lsif_data_docs_search_private_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					-- Decrement tally counting tables.
					UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
		},
	} {
		if normalizeFunction(testCase.lhs) != normalizeFunction(testCase.rhs) {
			t.Run(testCase.name, func(t *testing.T) {
				t.Errorf("unexpected comparison. %q != %q", testCase.lhs, testCase.rhs)
			})
		}
	}
}
