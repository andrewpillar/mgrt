-- mgrt: revision: 1136214247:
-- mgrt: author: test <test@example.com>
-- mgrt: up

CREATE TABLE third_table (
	id INTEGER PRIMARY KEY NOT NULL
);

-- mgrt: down

DROP TABLE third_table;
