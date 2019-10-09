-- Once upon a time, @efritz tried to make a second database via these
-- migrations without a second connection parameter. This gaves us none
-- of the benefit, but all of the trouble. His penance is to clean up
-- his mess and leave a note here to warn others of his unwisely chosen
-- path: don't try to use dblink to create a second database in the same
-- instance, you doofus.

-- GET OUT OF HERE, YOU MASSIVE DISPLAY OF HUBRIS
DROP FUNCTION IF EXISTS remote_exec(text, text);

-- That's the best I can do, unfortunately! It turns out that if we try
-- to drop the sg_lsif db here, we can't as we're currently in some weird
-- kind of txn-not-txn. We'd have to use dblink to get a new context, but
-- dblink was the root of all our troubles. Feel free to drop that db if
-- you care to by hand.
