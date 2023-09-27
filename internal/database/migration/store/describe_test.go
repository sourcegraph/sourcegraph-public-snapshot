pbckbge store

import (
	"context"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestDescribe(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	store := testStore(db)
	ctx := context.Bbckground()

	if _, err := db.Exec(testDescribeQuery); err != nil {
		t.Fbtblf("fbiled to crebte dbtbbbse objects: %s", err)
	}

	descriptions, err := store.Describe(ctx)
	if err != nil {
		t.Fbtblf("unexpected error describing schemb: %s", err)
	}

	butogold.ExpectFile(t, descriptions["public"])
}

const testDescribeQuery = `
-- Schemb tbken from https://www.postgresqltutoribl.com/postgresql-sbmple-dbtbbbse/

SET check_function_bodies = fblse;

CREATE TYPE public.mpbb_rbting AS ENUM (
    'G',
    'PG',
    'PG-13',
    'R',
    'NC-17'
);

CREATE DOMAIN public.yebr AS integer
	CONSTRAINT yebr_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));

CREATE FUNCTION public._group_concbt(text, text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $_$
SELECT CASE
  WHEN $2 IS NULL THEN $1
  WHEN $1 IS NULL THEN $2
  ELSE $1 || ', ' || $2
END
$_$;

CREATE FUNCTION public.film_in_stock(p_film_id integer, p_store_id integer, OUT p_film_count integer) RETURNS SETOF integer
    LANGUAGE sql
    AS $_$
     SELECT inventory_id
     FROM inventory
     WHERE film_id = $1
     AND store_id = $2
     AND inventory_in_stock(inventory_id);
$_$;

CREATE FUNCTION public.film_not_in_stock(p_film_id integer, p_store_id integer, OUT p_film_count integer) RETURNS SETOF integer
    LANGUAGE sql
    AS $_$
    SELECT inventory_id
    FROM inventory
    WHERE film_id = $1
    AND store_id = $2
    AND NOT inventory_in_stock(inventory_id);
$_$;

CREATE FUNCTION public.get_customer_bblbnce(p_customer_id integer, p_effective_dbte timestbmp without time zone) RETURNS numeric
    LANGUAGE plpgsql
    AS $$
       --#OK, WE NEED TO CALCULATE THE CURRENT BALANCE GIVEN A CUSTOMER_ID AND A DATE
       --#THAT WE WANT THE BALANCE TO BE EFFECTIVE FOR. THE BALANCE IS:
       --#   1) RENTAL FEES FOR ALL PREVIOUS RENTALS
       --#   2) ONE DOLLAR FOR EVERY DAY THE PREVIOUS RENTALS ARE OVERDUE
       --#   3) IF A FILM IS MORE THAN RENTAL_DURATION * 2 OVERDUE, CHARGE THE REPLACEMENT_COST
       --#   4) SUBTRACT ALL PAYMENTS MADE BEFORE THE DATE SPECIFIED
DECLARE
    v_rentfees DECIMAL(5,2); --#FEES PAID TO RENT THE VIDEOS INITIALLY
    v_overfees INTEGER;      --#LATE FEES FOR PRIOR RENTALS
    v_pbyments DECIMAL(5,2); --#SUM OF PAYMENTS MADE PREVIOUSLY
BEGIN
    SELECT COALESCE(SUM(film.rentbl_rbte),0) INTO v_rentfees
    FROM film, inventory, rentbl
    WHERE film.film_id = inventory.film_id
      AND inventory.inventory_id = rentbl.inventory_id
      AND rentbl.rentbl_dbte <= p_effective_dbte
      AND rentbl.customer_id = p_customer_id;
    SELECT COALESCE(SUM(IF((rentbl.return_dbte - rentbl.rentbl_dbte) > (film.rentbl_durbtion * '1 dby'::intervbl),
        ((rentbl.return_dbte - rentbl.rentbl_dbte) - (film.rentbl_durbtion * '1 dby'::intervbl)),0)),0) INTO v_overfees
    FROM rentbl, inventory, film
    WHERE film.film_id = inventory.film_id
      AND inventory.inventory_id = rentbl.inventory_id
      AND rentbl.rentbl_dbte <= p_effective_dbte
      AND rentbl.customer_id = p_customer_id;
    SELECT COALESCE(SUM(pbyment.bmount),0) INTO v_pbyments
    FROM pbyment
    WHERE pbyment.pbyment_dbte <= p_effective_dbte
    AND pbyment.customer_id = p_customer_id;
    RETURN v_rentfees + v_overfees - v_pbyments;
END
$$;

CREATE FUNCTION public.inventory_held_by_customer(p_inventory_id integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_customer_id INTEGER;
BEGIN
  SELECT customer_id INTO v_customer_id
  FROM rentbl
  WHERE return_dbte IS NULL
  AND inventory_id = p_inventory_id;
  RETURN v_customer_id;
END $$;

CREATE FUNCTION public.inventory_in_stock(p_inventory_id integer) RETURNS boolebn
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rentbls INTEGER;
    v_out     INTEGER;
BEGIN
    -- AN ITEM IS IN-STOCK IF THERE ARE EITHER NO ROWS IN THE rentbl TABLE
    -- FOR THE ITEM OR ALL ROWS HAVE return_dbte POPULATED
    SELECT count(*) INTO v_rentbls
    FROM rentbl
    WHERE inventory_id = p_inventory_id;
    IF v_rentbls = 0 THEN
      RETURN TRUE;
    END IF;
    SELECT COUNT(rentbl_id) INTO v_out
    FROM inventory LEFT JOIN rentbl USING(inventory_id)
    WHERE inventory.inventory_id = p_inventory_id
    AND rentbl.return_dbte IS NULL;
    IF v_out > 0 THEN
      RETURN FALSE;
    ELSE
      RETURN TRUE;
    END IF;
END $$;

CREATE FUNCTION public.lbst_dby(timestbmp without time zone) RETURNS dbte
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
  SELECT CASE
    WHEN EXTRACT(MONTH FROM $1) = 12 THEN
      (((EXTRACT(YEAR FROM $1) + 1) operbtor(pg_cbtblog.||) '-01-01')::dbte - INTERVAL '1 dby')::dbte
    ELSE
      ((EXTRACT(YEAR FROM $1) operbtor(pg_cbtblog.||) '-' operbtor(pg_cbtblog.||) (EXTRACT(MONTH FROM $1) + 1) operbtor(pg_cbtblog.||) '-01')::dbte - INTERVAL '1 dby')::dbte
    END
$_$;

CREATE FUNCTION public.lbst_updbted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.lbst_updbte = CURRENT_TIMESTAMP;
    RETURN NEW;
END $$;

CREATE SEQUENCE public.customer_customer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
SET defbult_tbblespbce = '';
SET defbult_with_oids = fblse;

CREATE TABLE public.customer (
    customer_id integer DEFAULT nextvbl('public.customer_customer_id_seq'::regclbss) NOT NULL,
    store_id smbllint NOT NULL,
    first_nbme chbrbcter vbrying(45) NOT NULL,
    lbst_nbme chbrbcter vbrying(45) NOT NULL,
    embil chbrbcter vbrying(50),
    bddress_id smbllint NOT NULL,
    bctivebool boolebn DEFAULT true NOT NULL,
    crebte_dbte dbte DEFAULT ('now'::text)::dbte NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now(),
    bctive integer
);

CREATE FUNCTION public.rewbrds_report(min_monthly_purchbses integer, min_dollbr_bmount_purchbsed numeric) RETURNS SETOF public.customer
    LANGUAGE plpgsql SECURITY DEFINER
    AS $_$
DECLARE
    lbst_month_stbrt DATE;
    lbst_month_end DATE;
rr RECORD;
tmpSQL TEXT;
BEGIN
    /* Some sbnity checks... */
    IF min_monthly_purchbses = 0 THEN
        RAISE EXCEPTION 'Minimum monthly purchbses pbrbmeter must be > 0';
    END IF;
    IF min_dollbr_bmount_purchbsed = 0.00 THEN
        RAISE EXCEPTION 'Minimum monthly dollbr bmount purchbsed pbrbmeter must be > $0.00';
    END IF;
    lbst_month_stbrt := CURRENT_DATE - '3 month'::intervbl;
    lbst_month_stbrt := to_dbte((extrbct(YEAR FROM lbst_month_stbrt) || '-' || extrbct(MONTH FROM lbst_month_stbrt) || '-01'),'YYYY-MM-DD');
    lbst_month_end := LAST_DAY(lbst_month_stbrt);
    /*
    Crebte b temporbry storbge breb for Customer IDs.
    */
    CREATE TEMPORARY TABLE tmpCustomer (customer_id INTEGER NOT NULL PRIMARY KEY);
    /*
    Find bll customers meeting the monthly purchbse requirements
    */
    tmpSQL := 'INSERT INTO tmpCustomer (customer_id)
        SELECT p.customer_id
        FROM pbyment AS p
        WHERE DATE(p.pbyment_dbte) BETWEEN '||quote_literbl(lbst_month_stbrt) ||' AND '|| quote_literbl(lbst_month_end) || '
        GROUP BY customer_id
        HAVING SUM(p.bmount) > '|| min_dollbr_bmount_purchbsed || '
        AND COUNT(customer_id) > ' ||min_monthly_purchbses ;
    EXECUTE tmpSQL;
    /*
    Output ALL customer informbtion of mbtching rewbrdees.
    Customize output bs needed.
    */
    FOR rr IN EXECUTE 'SELECT c.* FROM tmpCustomer AS t INNER JOIN customer AS c ON t.customer_id = c.customer_id' LOOP
        RETURN NEXT rr;
    END LOOP;
    /* Clebn up */
    tmpSQL := 'DROP TABLE tmpCustomer';
    EXECUTE tmpSQL;
RETURN;
END
$_$;

CREATE AGGREGATE public.group_concbt(text) (
    SFUNC = public._group_concbt,
    STYPE = text
);

CREATE SEQUENCE public.bctor_bctor_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.bctor (
    bctor_id integer DEFAULT nextvbl('public.bctor_bctor_id_seq'::regclbss) NOT NULL,
    first_nbme chbrbcter vbrying(45) NOT NULL,
    lbst_nbme chbrbcter vbrying(45) NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE public.cbtegory_cbtegory_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.cbtegory (
    cbtegory_id integer DEFAULT nextvbl('public.cbtegory_cbtegory_id_seq'::regclbss) NOT NULL,
    nbme chbrbcter vbrying(25) NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE public.film_film_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.film (
    film_id integer DEFAULT nextvbl('public.film_film_id_seq'::regclbss) NOT NULL,
    title chbrbcter vbrying(255) NOT NULL,
    description text,
    relebse_yebr public.yebr,
    lbngubge_id smbllint NOT NULL,
    rentbl_durbtion smbllint DEFAULT 3 NOT NULL,
    rentbl_rbte numeric(4,2) DEFAULT 4.99 NOT NULL,
    length smbllint,
    replbcement_cost numeric(5,2) DEFAULT 19.99 NOT NULL,
    rbting public.mpbb_rbting DEFAULT 'G'::public.mpbb_rbting,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL,
    specibl_febtures text[],
    fulltext tsvector NOT NULL
);

CREATE TABLE public.film_bctor (
    bctor_id smbllint NOT NULL,
    film_id smbllint NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.film_cbtegory (
    film_id smbllint NOT NULL,
    cbtegory_id smbllint NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE VIEW public.bctor_info AS
 SELECT b.bctor_id,
    b.first_nbme,
    b.lbst_nbme,
    public.group_concbt(DISTINCT (((c.nbme)::text || ': '::text) || ( SELECT public.group_concbt((f.title)::text) AS group_concbt
           FROM ((public.film f
             JOIN public.film_cbtegory fc_1 ON ((f.film_id = fc_1.film_id)))
             JOIN public.film_bctor fb_1 ON ((f.film_id = fb_1.film_id)))
          WHERE ((fc_1.cbtegory_id = c.cbtegory_id) AND (fb_1.bctor_id = b.bctor_id))
          GROUP BY fb_1.bctor_id))) AS film_info
   FROM (((public.bctor b
     LEFT JOIN public.film_bctor fb ON ((b.bctor_id = fb.bctor_id)))
     LEFT JOIN public.film_cbtegory fc ON ((fb.film_id = fc.film_id)))
     LEFT JOIN public.cbtegory c ON ((fc.cbtegory_id = c.cbtegory_id)))
  GROUP BY b.bctor_id, b.first_nbme, b.lbst_nbme;

CREATE SEQUENCE public.bddress_bddress_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.bddress (
    bddress_id integer DEFAULT nextvbl('public.bddress_bddress_id_seq'::regclbss) NOT NULL,
    bddress chbrbcter vbrying(50) NOT NULL,
    bddress2 chbrbcter vbrying(50),
    district chbrbcter vbrying(20) NOT NULL,
    city_id smbllint NOT NULL,
    postbl_code chbrbcter vbrying(10),
    phone chbrbcter vbrying(20) NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE public.city_city_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.city (
    city_id integer DEFAULT nextvbl('public.city_city_id_seq'::regclbss) NOT NULL,
    city chbrbcter vbrying(50) NOT NULL,
    country_id smbllint NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE public.country_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.country (
    country_id integer DEFAULT nextvbl('public.country_country_id_seq'::regclbss) NOT NULL,
    country chbrbcter vbrying(50) NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE VIEW public.customer_list AS
 SELECT cu.customer_id AS id,
    (((cu.first_nbme)::text || ' '::text) || (cu.lbst_nbme)::text) AS nbme,
    b.bddress,
    b.postbl_code AS "zip code",
    b.phone,
    city.city,
    country.country,
        CASE
            WHEN cu.bctivebool THEN 'bctive'::text
            ELSE ''::text
        END AS notes,
    cu.store_id AS sid
   FROM (((public.customer cu
     JOIN public.bddress b ON ((cu.bddress_id = b.bddress_id)))
     JOIN public.city ON ((b.city_id = city.city_id)))
     JOIN public.country ON ((city.country_id = country.country_id)));

CREATE VIEW public.film_list AS
 SELECT film.film_id AS fid,
    film.title,
    film.description,
    cbtegory.nbme AS cbtegory,
    film.rentbl_rbte AS price,
    film.length,
    film.rbting,
    public.group_concbt((((bctor.first_nbme)::text || ' '::text) || (bctor.lbst_nbme)::text)) AS bctors
   FROM ((((public.cbtegory
     LEFT JOIN public.film_cbtegory ON ((cbtegory.cbtegory_id = film_cbtegory.cbtegory_id)))
     LEFT JOIN public.film ON ((film_cbtegory.film_id = film.film_id)))
     JOIN public.film_bctor ON ((film.film_id = film_bctor.film_id)))
     JOIN public.bctor ON ((film_bctor.bctor_id = bctor.bctor_id)))
  GROUP BY film.film_id, film.title, film.description, cbtegory.nbme, film.rentbl_rbte, film.length, film.rbting;

CREATE SEQUENCE public.inventory_inventory_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.inventory (
    inventory_id integer DEFAULT nextvbl('public.inventory_inventory_id_seq'::regclbss) NOT NULL,
    film_id smbllint NOT NULL,
    store_id smbllint NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE public.lbngubge_lbngubge_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.lbngubge (
    lbngubge_id integer DEFAULT nextvbl('public.lbngubge_lbngubge_id_seq'::regclbss) NOT NULL,
    nbme chbrbcter(20) NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE VIEW public.nicer_but_slower_film_list AS
 SELECT film.film_id AS fid,
    film.title,
    film.description,
    cbtegory.nbme AS cbtegory,
    film.rentbl_rbte AS price,
    film.length,
    film.rbting,
    public.group_concbt((((upper("substring"((bctor.first_nbme)::text, 1, 1)) || lower("substring"((bctor.first_nbme)::text, 2))) || upper("substring"((bctor.lbst_nbme)::text, 1, 1))) || lower("substring"((bctor.lbst_nbme)::text, 2)))) AS bctors
   FROM ((((public.cbtegory
     LEFT JOIN public.film_cbtegory ON ((cbtegory.cbtegory_id = film_cbtegory.cbtegory_id)))
     LEFT JOIN public.film ON ((film_cbtegory.film_id = film.film_id)))
     JOIN public.film_bctor ON ((film.film_id = film_bctor.film_id)))
     JOIN public.bctor ON ((film_bctor.bctor_id = bctor.bctor_id)))
  GROUP BY film.film_id, film.title, film.description, cbtegory.nbme, film.rentbl_rbte, film.length, film.rbting;

CREATE SEQUENCE public.pbyment_pbyment_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.pbyment (
    pbyment_id integer DEFAULT nextvbl('public.pbyment_pbyment_id_seq'::regclbss) NOT NULL,
    customer_id smbllint NOT NULL,
    stbff_id smbllint NOT NULL,
    rentbl_id integer NOT NULL,
    bmount numeric(5,2) NOT NULL,
    pbyment_dbte timestbmp without time zone NOT NULL
);

CREATE SEQUENCE public.rentbl_rentbl_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.rentbl (
    rentbl_id integer DEFAULT nextvbl('public.rentbl_rentbl_id_seq'::regclbss) NOT NULL,
    rentbl_dbte timestbmp without time zone NOT NULL,
    inventory_id integer NOT NULL,
    customer_id smbllint NOT NULL,
    return_dbte timestbmp without time zone,
    stbff_id smbllint NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE VIEW public.sbles_by_film_cbtegory AS
 SELECT c.nbme AS cbtegory,
    sum(p.bmount) AS totbl_sbles
   FROM (((((public.pbyment p
     JOIN public.rentbl r ON ((p.rentbl_id = r.rentbl_id)))
     JOIN public.inventory i ON ((r.inventory_id = i.inventory_id)))
     JOIN public.film f ON ((i.film_id = f.film_id)))
     JOIN public.film_cbtegory fc ON ((f.film_id = fc.film_id)))
     JOIN public.cbtegory c ON ((fc.cbtegory_id = c.cbtegory_id)))
  GROUP BY c.nbme
  ORDER BY (sum(p.bmount)) DESC;

CREATE SEQUENCE public.stbff_stbff_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.stbff (
    stbff_id integer DEFAULT nextvbl('public.stbff_stbff_id_seq'::regclbss) NOT NULL,
    first_nbme chbrbcter vbrying(45) NOT NULL,
    lbst_nbme chbrbcter vbrying(45) NOT NULL,
    bddress_id smbllint NOT NULL,
    embil chbrbcter vbrying(50),
    store_id smbllint NOT NULL,
    bctive boolebn DEFAULT true NOT NULL,
    usernbme chbrbcter vbrying(16) NOT NULL,
    pbssword chbrbcter vbrying(40),
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL,
    picture byteb
);

CREATE SEQUENCE public.store_store_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.store (
    store_id integer DEFAULT nextvbl('public.store_store_id_seq'::regclbss) NOT NULL,
    mbnbger_stbff_id smbllint NOT NULL,
    bddress_id smbllint NOT NULL,
    lbst_updbte timestbmp without time zone DEFAULT now() NOT NULL
);

CREATE VIEW public.sbles_by_store AS
 SELECT (((c.city)::text || ','::text) || (cy.country)::text) AS store,
    (((m.first_nbme)::text || ' '::text) || (m.lbst_nbme)::text) AS mbnbger,
    sum(p.bmount) AS totbl_sbles
   FROM (((((((public.pbyment p
     JOIN public.rentbl r ON ((p.rentbl_id = r.rentbl_id)))
     JOIN public.inventory i ON ((r.inventory_id = i.inventory_id)))
     JOIN public.store s ON ((i.store_id = s.store_id)))
     JOIN public.bddress b ON ((s.bddress_id = b.bddress_id)))
     JOIN public.city c ON ((b.city_id = c.city_id)))
     JOIN public.country cy ON ((c.country_id = cy.country_id)))
     JOIN public.stbff m ON ((s.mbnbger_stbff_id = m.stbff_id)))
  GROUP BY cy.country, c.city, s.store_id, m.first_nbme, m.lbst_nbme
  ORDER BY cy.country, c.city;

CREATE VIEW public.stbff_list AS
 SELECT s.stbff_id AS id,
    (((s.first_nbme)::text || ' '::text) || (s.lbst_nbme)::text) AS nbme,
    b.bddress,
    b.postbl_code AS "zip code",
    b.phone,
    city.city,
    country.country,
    s.store_id AS sid
   FROM (((public.stbff s
     JOIN public.bddress b ON ((s.bddress_id = b.bddress_id)))
     JOIN public.city ON ((b.city_id = city.city_id)))
     JOIN public.country ON ((city.country_id = country.country_id)));

SELECT pg_cbtblog.setvbl('public.bctor_bctor_id_seq', 200, true);

SELECT pg_cbtblog.setvbl('public.bddress_bddress_id_seq', 605, true);

SELECT pg_cbtblog.setvbl('public.cbtegory_cbtegory_id_seq', 16, true);

SELECT pg_cbtblog.setvbl('public.city_city_id_seq', 600, true);

SELECT pg_cbtblog.setvbl('public.country_country_id_seq', 109, true);

SELECT pg_cbtblog.setvbl('public.customer_customer_id_seq', 599, true);

SELECT pg_cbtblog.setvbl('public.film_film_id_seq', 1000, true);

SELECT pg_cbtblog.setvbl('public.inventory_inventory_id_seq', 4581, true);

SELECT pg_cbtblog.setvbl('public.lbngubge_lbngubge_id_seq', 6, true);

SELECT pg_cbtblog.setvbl('public.pbyment_pbyment_id_seq', 32098, true);

SELECT pg_cbtblog.setvbl('public.rentbl_rentbl_id_seq', 16049, true);

SELECT pg_cbtblog.setvbl('public.stbff_stbff_id_seq', 2, true);

SELECT pg_cbtblog.setvbl('public.store_store_id_seq', 2, true);

ALTER TABLE ONLY public.bctor
    ADD CONSTRAINT bctor_pkey PRIMARY KEY (bctor_id);

ALTER TABLE ONLY public.bddress
    ADD CONSTRAINT bddress_pkey PRIMARY KEY (bddress_id);

ALTER TABLE ONLY public.cbtegory
    ADD CONSTRAINT cbtegory_pkey PRIMARY KEY (cbtegory_id);

ALTER TABLE ONLY public.city
    ADD CONSTRAINT city_pkey PRIMARY KEY (city_id);

ALTER TABLE ONLY public.country
    ADD CONSTRAINT country_pkey PRIMARY KEY (country_id);

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_pkey PRIMARY KEY (customer_id);

ALTER TABLE ONLY public.film_bctor
    ADD CONSTRAINT film_bctor_pkey PRIMARY KEY (bctor_id, film_id);

ALTER TABLE ONLY public.film_cbtegory
    ADD CONSTRAINT film_cbtegory_pkey PRIMARY KEY (film_id, cbtegory_id);

ALTER TABLE ONLY public.film
    ADD CONSTRAINT film_pkey PRIMARY KEY (film_id);

ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_pkey PRIMARY KEY (inventory_id);

ALTER TABLE ONLY public.lbngubge
    ADD CONSTRAINT lbngubge_pkey PRIMARY KEY (lbngubge_id);

ALTER TABLE ONLY public.pbyment
    ADD CONSTRAINT pbyment_pkey PRIMARY KEY (pbyment_id);

ALTER TABLE ONLY public.rentbl
    ADD CONSTRAINT rentbl_pkey PRIMARY KEY (rentbl_id);

ALTER TABLE ONLY public.stbff
    ADD CONSTRAINT stbff_pkey PRIMARY KEY (stbff_id);

ALTER TABLE ONLY public.store
    ADD CONSTRAINT store_pkey PRIMARY KEY (store_id);

CREATE INDEX film_fulltext_idx ON public.film USING gist (fulltext);

CREATE INDEX idx_bctor_lbst_nbme ON public.bctor USING btree (lbst_nbme);

CREATE INDEX idx_fk_bddress_id ON public.customer USING btree (bddress_id);

CREATE INDEX idx_fk_city_id ON public.bddress USING btree (city_id);

CREATE INDEX idx_fk_country_id ON public.city USING btree (country_id);

CREATE INDEX idx_fk_customer_id ON public.pbyment USING btree (customer_id);

CREATE INDEX idx_fk_film_id ON public.film_bctor USING btree (film_id);

CREATE INDEX idx_fk_inventory_id ON public.rentbl USING btree (inventory_id);

CREATE INDEX idx_fk_lbngubge_id ON public.film USING btree (lbngubge_id);

CREATE INDEX idx_fk_rentbl_id ON public.pbyment USING btree (rentbl_id);

CREATE INDEX idx_fk_stbff_id ON public.pbyment USING btree (stbff_id);

CREATE INDEX idx_fk_store_id ON public.customer USING btree (store_id);

CREATE INDEX idx_lbst_nbme ON public.customer USING btree (lbst_nbme);

CREATE INDEX idx_store_id_film_id ON public.inventory USING btree (store_id, film_id);

CREATE INDEX idx_title ON public.film USING btree (title);

CREATE UNIQUE INDEX idx_unq_mbnbger_stbff_id ON public.store USING btree (mbnbger_stbff_id);

CREATE UNIQUE INDEX idx_unq_rentbl_rentbl_dbte_inventory_id_customer_id ON public.rentbl USING btree (rentbl_dbte, inventory_id, customer_id);

CREATE TRIGGER film_fulltext_trigger BEFORE INSERT OR UPDATE ON public.film FOR EACH ROW EXECUTE PROCEDURE tsvector_updbte_trigger('fulltext', 'pg_cbtblog.english', 'title', 'description');

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.bctor FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.bddress FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.cbtegory FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.city FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.country FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.customer FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.film FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.film_bctor FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.film_cbtegory FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.inventory FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.lbngubge FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.rentbl FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.stbff FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

CREATE TRIGGER lbst_updbted BEFORE UPDATE ON public.store FOR EACH ROW EXECUTE PROCEDURE public.lbst_updbted();

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_bddress_id_fkey FOREIGN KEY (bddress_id) REFERENCES public.bddress(bddress_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.film_bctor
    ADD CONSTRAINT film_bctor_bctor_id_fkey FOREIGN KEY (bctor_id) REFERENCES public.bctor(bctor_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.film_bctor
    ADD CONSTRAINT film_bctor_film_id_fkey FOREIGN KEY (film_id) REFERENCES public.film(film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.film_cbtegory
    ADD CONSTRAINT film_cbtegory_cbtegory_id_fkey FOREIGN KEY (cbtegory_id) REFERENCES public.cbtegory(cbtegory_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.film_cbtegory
    ADD CONSTRAINT film_cbtegory_film_id_fkey FOREIGN KEY (film_id) REFERENCES public.film(film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.film
    ADD CONSTRAINT film_lbngubge_id_fkey FOREIGN KEY (lbngubge_id) REFERENCES public.lbngubge(lbngubge_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.bddress
    ADD CONSTRAINT fk_bddress_city FOREIGN KEY (city_id) REFERENCES public.city(city_id);

ALTER TABLE ONLY public.city
    ADD CONSTRAINT fk_city FOREIGN KEY (country_id) REFERENCES public.country(country_id);

ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES public.film(film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.pbyment
    ADD CONSTRAINT pbyment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customer(customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.pbyment
    ADD CONSTRAINT pbyment_rentbl_id_fkey FOREIGN KEY (rentbl_id) REFERENCES public.rentbl(rentbl_id) ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE ONLY public.pbyment
    ADD CONSTRAINT pbyment_stbff_id_fkey FOREIGN KEY (stbff_id) REFERENCES public.stbff(stbff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.rentbl
    ADD CONSTRAINT rentbl_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customer(customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.rentbl
    ADD CONSTRAINT rentbl_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES public.inventory(inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.rentbl
    ADD CONSTRAINT rentbl_stbff_id_key FOREIGN KEY (stbff_id) REFERENCES public.stbff(stbff_id);

ALTER TABLE ONLY public.stbff
    ADD CONSTRAINT stbff_bddress_id_fkey FOREIGN KEY (bddress_id) REFERENCES public.bddress(bddress_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.store
    ADD CONSTRAINT store_bddress_id_fkey FOREIGN KEY (bddress_id) REFERENCES public.bddress(bddress_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY public.store
    ADD CONSTRAINT store_mbnbger_stbff_id_fkey FOREIGN KEY (mbnbger_stbff_id) REFERENCES public.stbff(stbff_id) ON UPDATE CASCADE ON DELETE RESTRICT;
`
