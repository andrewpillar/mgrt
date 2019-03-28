-- mgrt: revision: 1136214248:
-- mgrt: author: test <test@example.com>
-- mgrt: up

CREATE TABLE fourth_table (
	id INTEGER PRIMARY KEY NOT NULL
);

-- mgrt: down

DROP TABLE fourth_table;
