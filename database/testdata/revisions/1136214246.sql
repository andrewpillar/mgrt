-- mgrt: revision: 1136214246:
-- mgrt: author: test <test@example.com>
-- mgrt: up

ALTER TABLE example ADD COLUMN example_int INTEGER;

-- mgrt: down

ALTER TABLE example RENAME example_int TO _example_int;
