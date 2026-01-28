package internal

import (
	"testing"

	"github.com/gnolang/supernova/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validBaseConfig returns a config with all fields valid except Mode/MixRatio
func validBaseConfig() *Config {
	return &Config{
		URL:     "http://localhost:26657",
		ChainID: "dev",
		Mnemonic: "source bonus chronic canvas draft south burst lottery " +
			"vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast",
		SubAccounts:  10,
		Transactions: 100,
		BatchSize:    100,
	}
}

func TestConfig_Validate_MixedModeRequiresRatio(t *testing.T) {
	t.Parallel()

	cfg := validBaseConfig()
	cfg.Mode = runtime.Mixed.String()
	cfg.MixRatio = ""

	err := cfg.Validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, errMixRatioRequired)
}

func TestConfig_Validate_MixedModeInvalidRatio(t *testing.T) {
	t.Parallel()

	cfg := validBaseConfig()
	cfg.Mode = runtime.Mixed.String()
	cfg.MixRatio = "REALM_CALL:50" // only one type, needs at least 2

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mix-ratio")
}

func TestConfig_Validate_MixedModeValidRatio(t *testing.T) {
	t.Parallel()

	cfg := validBaseConfig()
	cfg.Mode = runtime.Mixed.String()
	cfg.MixRatio = "REALM_CALL:70,REALM_DEPLOYMENT:30"

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_NonMixedModeIgnoresRatio(t *testing.T) {
	t.Parallel()

	cfg := validBaseConfig()
	cfg.Mode = runtime.RealmCall.String()
	cfg.MixRatio = "this is invalid but should be ignored"

	err := cfg.Validate()
	assert.NoError(t, err)
}
