-- mgrt: revision: 1136214245:
-- mgrt: author: test <test@example.com>
-- mgrt: up

CREATE TABLE example (
	id INTEGER PRIMARY KEY NOT NULL
);

-- mgrt: down

DROP TABLE example;
