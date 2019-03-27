-- mgrt: revision: 1136214248:
-- mgrt: author: test <test@example.com>
-- mgrt: up

ALTER TABLE example ADD COLUMN example_timestamp TIMESTAMP;

-- mgrt: down

ALTER TABLE example RENAME example_timestamp TO _example_timestamp;
