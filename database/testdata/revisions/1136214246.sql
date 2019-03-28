-- mgrt: revision: 1136214246:
-- mgrt: author: test <test@example.com>
-- mgrt: up

CREATE TABLE second_table (
	id INTEGER PRIMARY KEY NOT NULL
);

-- mgrt: down

DROP TABLE second_table;
