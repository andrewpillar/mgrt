package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	if err := Create(); err != nil {
		t.Errorf("failed to create config: %s\n", err)
	}

	bactual, err := ioutil.ReadFile(file)

	if err != nil {
		t.Errorf("failed to read config: %s\n", err)
	}

	bexpected := []byte(stub)

	if len(bactual) != len(bexpected) {
		t.Errorf("actual config file does not match expected: mismatched lengths: expected = '%d', actual = '%d'\n", len(bexpected), len(bactual))
	}

	for i := range bactual {
		if bactual[i] != bexpected[i] {
			t.Errorf("actual config file does not match expected: mismatched characters: expected = '%s', actual = '%s'\n", string(bexpected[i]), string(bactual[i]))
		}
	}

	cfg, err := Open()

	if err != nil {
		t.Errorf("failed to open config: %s\n", err)
	}

	cfg.Close()

	os.RemoveAll(file)
}
