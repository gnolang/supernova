package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMixRatio_Valid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		input    string
		expected []MixRatio
	}{
		{
			"three-way split",
			"REALM_CALL:70,REALM_DEPLOYMENT:20,PACKAGE_DEPLOYMENT:10",
			[]MixRatio{
				{RealmCall, 70},
				{RealmDeployment, 20},
				{PackageDeployment, 10},
			},
		},
		{
			"two-way split",
			"REALM_CALL:50,REALM_DEPLOYMENT:50",
			[]MixRatio{
				{RealmCall, 50},
				{RealmDeployment, 50},
			},
		},
		{
			"with spaces",
			"REALM_CALL:70, REALM_DEPLOYMENT:30",
			[]MixRatio{
				{RealmCall, 70},
				{RealmDeployment, 30},
			},
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := ParseMixRatio(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, config.Ratios)
		})
	}
}

func TestParseMixRatio_Invalid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			"empty string",
			"",
			errEmptyMixRatio,
		},
		{
			"whitespace only",
			"   ",
			errEmptyMixRatio,
		},
		{
			"single type",
			"REALM_CALL:100",
			errInsufficientTypes,
		},
		{
			"missing colon",
			"REALM_CALL70,REALM_DEPLOYMENT30",
			errInvalidRatioFormat,
		},
		{
			"invalid percentage - negative",
			"REALM_CALL:-10,REALM_DEPLOYMENT:110",
			errInvalidPercentage,
		},
		{
			"invalid percentage - zero",
			"REALM_CALL:0,REALM_DEPLOYMENT:100",
			errInvalidPercentage,
		},
		{
			"invalid percentage - over 100",
			"REALM_CALL:101,REALM_DEPLOYMENT:0",
			errInvalidPercentage,
		},
		{
			"invalid percentage - not a number",
			"REALM_CALL:abc,REALM_DEPLOYMENT:50",
			errInvalidPercentage,
		},
		{
			"unknown type",
			"UNKNOWN_TYPE:50,REALM_DEPLOYMENT:50",
			errUnknownType,
		},
		{
			"duplicate type",
			"REALM_CALL:50,REALM_CALL:50",
			errDuplicateType,
		},
		{
			"MIXED in mix",
			"MIXED:50,REALM_CALL:50",
			errMixedInMix,
		},
		{
			"sum not 100 - under",
			"REALM_CALL:40,REALM_DEPLOYMENT:40",
			errRatioSumNot100,
		},
		{
			"sum not 100 - over",
			"REALM_CALL:60,REALM_DEPLOYMENT:60",
			errRatioSumNot100,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseMixRatio(tc.input)
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestMixConfig_HasType(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []MixRatio{
			{RealmCall, 70},
			{RealmDeployment, 30},
		},
	}

	assert.True(t, config.HasType(RealmCall))
	assert.True(t, config.HasType(RealmDeployment))
	assert.False(t, config.HasType(PackageDeployment))
	assert.False(t, config.HasType(Mixed))
}

func TestMixConfig_CalculateTransactionCounts(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		ratios   []MixRatio
		total    uint64
		expected map[Type]uint64
	}{
		{
			"exact division",
			[]MixRatio{
				{RealmCall, 70},
				{RealmDeployment, 20},
				{PackageDeployment, 10},
			},
			100,
			map[Type]uint64{
				RealmCall:         70,
				RealmDeployment:   20,
				PackageDeployment: 10,
			},
		},
		{
			"with rounding - remainder goes to last",
			[]MixRatio{
				{RealmCall, 70},
				{RealmDeployment, 30},
			},
			10,
			map[Type]uint64{
				RealmCall:       7,
				RealmDeployment: 3,
			},
		},
		{
			"small total with rounding",
			[]MixRatio{
				{RealmCall, 33},
				{RealmDeployment, 33},
				{PackageDeployment, 34},
			},
			10,
			map[Type]uint64{
				RealmCall:         3,
				RealmDeployment:   3,
				PackageDeployment: 4,
			},
		},
		{
			"two-way 50/50",
			[]MixRatio{
				{RealmCall, 50},
				{PackageDeployment, 50},
			},
			100,
			map[Type]uint64{
				RealmCall:         50,
				PackageDeployment: 50,
			},
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := &MixConfig{Ratios: tc.ratios}
			counts := config.CalculateTransactionCounts(tc.total)
			assert.Equal(t, tc.expected, counts)
		})
	}
}
