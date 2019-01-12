package database

func (db *DB) initSqlite3() error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = $1")

	if err != nil {
		return err
	}

	defer stmt.Close()

	var count int

	row := stmt.QueryRow(db.table)

	if err := row.Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return ErrInitialized
	}

	_, err = db.Exec(`
		CREATE TABLE ` + db.table + ` (
			id          INTEGER NOT NULL,
			hash        BLOB NOT NULL,
			direction   INTEGER NOT NULL,
			created_at  TIMESTAMP NOT NULL
		);
	`)

	return err
}
