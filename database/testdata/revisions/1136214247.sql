-- mgrt: revision: 1136214247:
-- mgrt: author: test <test@example.com>
-- mgrt: up

ALTER TABLE example ADD COLUMN example_text TEXT;

-- mgrt: down

ALTER TABLE example RENAME example_text TO _example_text;
