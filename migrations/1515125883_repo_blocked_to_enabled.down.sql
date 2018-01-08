UPDATE repo SET blocked = (not enabled);
ALTER TABLE repo DROP COLUMN enabled;
