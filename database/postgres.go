package database

func (db *DB) initPostgres() error {
	_, err := db.Exec(`
		CREATE TABLE mgrt_revisions (
			id         INT NOT NULL,
			hash       BYTEA NOT NULL,
			direction  INT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`)

	if err != nil {

	}

	return nil
}
