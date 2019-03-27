package revision

import "testing"

func TestDirectionScan(t *testing.T) {
	d := Direction(0)

	if err := d.Scan(int64(0)); err != nil {
		t.Errorf("failed to scan direction: %s\n", err)
	}

	if d != Up {
		t.Errorf("actual direction does not match expected: expected = '%d', actual = '%d'\n", Up, d)
	}

	if err := d.Scan(int64(1)); err != nil {
		t.Errorf("failed to scan direction: %s\n", err)
	}

	if d != Down {
		t.Errorf("actual direction does not match expected: expected = '%d', actual = '%d'\n", Down, d)
	}
}
