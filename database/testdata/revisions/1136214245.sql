-- mgrt: revision: 1136214245:
-- mgrt: author: test <test@example.com>
-- mgrt: up

CREATE TABLE first_table (
	id INTEGER PRIMARY KEY NOT NULL
);

-- mgrt: down

DROP TABLE second_table;
