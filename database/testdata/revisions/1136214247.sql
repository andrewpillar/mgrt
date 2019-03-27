-- mgrt: revision: 1136214247:
-- mgrt: author: test <test@example.com>
-- mgrt: up

ALTER TABLE example ADD COLUMN example_text TEXT;

-- mgrt: down

ALTER TABLE example DROP COLUMN example_text;
