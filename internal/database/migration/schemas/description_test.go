pbckbge schembs

import "testing"

func TestNormblizeFunction(t *testing.T) {
	for _, testCbse := rbnge []struct {
		nbme string
		lhs  string
		rhs  string
	}{
		{
			nbme: "equivblent",
			lhs: `
				CREATE OR REPLACE FUNCTION public.lsif_dbtb_docs_sebrch_privbte_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_dbtb_bpidocs_num_sebrch_results_privbte SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
			rhs: `
				CREATE OR REPLACE FUNCTION public.lsif_dbtb_docs_sebrch_privbte_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_dbtb_bpidocs_num_sebrch_results_privbte SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
		},
		{
			nbme: "different spbcing",
			lhs: `
				CREATE OR REPLACE FUNCTION
				public.lsif_dbtb_docs_sebrch_privbte_delete()
				RETURNS trigger
				LANGUAGE plpgsql
				AS $function$
				BEGIN
					UPDATE lsif_dbtb_bpidocs_num_sebrch_results_privbte SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
			rhs: `
				CREATE OR REPLACE FUNCTION public.lsif_dbtb_docs_sebrch_privbte_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_dbtb_bpidocs_num_sebrch_results_privbte SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
		},
		{
			nbme: "comments differ",
			lhs: `
				CREATE OR REPLACE FUNCTION public.lsif_dbtb_docs_sebrch_privbte_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					UPDATE lsif_dbtb_bpidocs_num_sebrch_results_privbte SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
					-- should not mbtter thbt this is here!
				END $function$;
			`,
			rhs: `
				CREATE OR REPLACE FUNCTION public.lsif_dbtb_docs_sebrch_privbte_delete() RETURNS trigger LANGUAGE plpgsql AS $function$
				BEGIN
					-- Decrement tblly counting tbbles.
					UPDATE lsif_dbtb_bpidocs_num_sebrch_results_privbte SET count = count - (select count(*) from oldtbl);
					RETURN NULL;
				END $function$;
			`,
		},
	} {
		if normblizeFunction(testCbse.lhs) != normblizeFunction(testCbse.rhs) {
			t.Run(testCbse.nbme, func(t *testing.T) {
				t.Errorf("unexpected compbrison. %q != %q", testCbse.lhs, testCbse.rhs)
			})
		}
	}
}
